package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

var (
	// Core styles
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))            // dark green
	success2Style = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))             // green
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))             // red
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))            // yellow
	pendingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // blue
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))            // cyan
	debugStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))           // light grey
	detailStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))            // purple
	streamStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))           // grey
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")) // purple

	// Additional config
	basePadding = 2
)

var StyleSymbols = map[string]string{
	"pass":    "✓",
	"fail":    "✗",
	"warning": "!",
	"pending": "◉",
	"info":    "ℹ",
	"arrow":   "→",
	"bullet":  "•",
	"dot":     "·",
	"hline":   "━",
}

func PrintSuccess(text string) {
	fmt.Println(successStyle.Render(text))
}
func PrintSuccess2(text string) {
	fmt.Println(success2Style.Render(text))
}
func PrintError(text string) {
	fmt.Println(errorStyle.Render(text))
}
func PrintWarning(text string) {
	fmt.Println(warningStyle.Render(text))
}
func PrintInfo(text string) {
	fmt.Println(infoStyle.Render(text))
}
func PrintDebug(text string) {
	fmt.Println(debugStyle.Render(text))
}
func PrintDetail(text string) {
	fmt.Println(detailStyle.Render(text))
}
func PrintStream(text string) {
	fmt.Println(streamStyle.Render(text))
}
func PrintHeader(text string) {
	fmt.Println(headerStyle.Render(text))
}
func FSuccess(text string) string {
	return successStyle.Render(text)
}
func FSuccess2(text string) string {
	return success2Style.Render(text)
}
func FError(text string) string {
	return errorStyle.Render(text)
}
func FWarning(text string) string {
	return warningStyle.Render(text)
}
func FInfo(text string) string {
	return infoStyle.Render(text)
}
func FDebug(text string) string {
	return debugStyle.Render(text)
}
func FDetail(text string) string {
	return detailStyle.Render(text)
}
func FStream(text string) string {
	return streamStyle.Render(text)
}
func FHeader(text string) string {
	return headerStyle.Render(text)
}

// ======================================== =================
// ======================================== Table Definitions
// ======================================== =================

type Table struct {
	Headers []string
	Rows    [][]string
	table   *table.Table
}

func NewTable(headers []string) *Table {
	t := &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
	t.table = table.New().Headers(headers...)
	t.table = t.table.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Padding(0, 1)
		}
		return lipgloss.NewStyle().Padding(0, 1)
	})
	return t
}

func (t *Table) ReconcileRows() {
	if len(t.Rows) == 0 {
		return
	}
	for _, row := range t.Rows {
		t.table.Row(row...)
	}
}

func (t *Table) FormatTable(useMarkdown bool) string {
	t.ReconcileRows()
	if useMarkdown {
		return t.table.Border(lipgloss.MarkdownBorder()).String()
	}
	return t.table.String()
}

func (t *Table) PrintTable(useMarkdown bool) {
	fmt.Println(t.FormatTable(useMarkdown))
}

func (t *Table) WriteMarkdownTableToFile(outputPath string) error {
	return os.WriteFile(outputPath, []byte(t.FormatTable(true)), 0644)
}

// =========================================== ==============
// =========================================== Output Manager
// =========================================== ==============

type FunctionOutput struct {
	Name        string
	Status      string
	Message     string
	StreamLines []string
	Complete    bool
	StartTime   time.Time
	LastUpdated time.Time
	Error       error
	Tables      map[string]*Table // Function tables
	Index       int
}

type ErrorReport struct {
	FunctionName string
	Error        error
	Time         time.Time
}

// Output manager main structure
type Manager struct {
	outputs         map[string]*FunctionOutput
	mutex           sync.RWMutex
	numLines        int
	maxStreams      int               // Max output stream lines per function
	unlimitedOutput bool              // When true, unlimited output per function
	tables          map[string]*Table // Global tables
	errors          []ErrorReport
	doneCh          chan struct{} // Channel to signal stopping the display
	pauseCh         chan bool     // Channel to pause/resume display updates
	isPaused        bool
	displayTick     time.Duration // Interval between display updates
	functionCount   int
	displayWg       sync.WaitGroup // WaitGroup for display goroutine shutdown
}

func NewManager(maxStreams int) *Manager {
	if maxStreams <= 0 {
		maxStreams = 15 // Default
	}
	return &Manager{
		outputs:         make(map[string]*FunctionOutput),
		tables:          make(map[string]*Table),
		errors:          []ErrorReport{},
		maxStreams:      maxStreams,
		unlimitedOutput: false,
		doneCh:          make(chan struct{}),
		pauseCh:         make(chan bool),
		isPaused:        false,
		displayTick:     200 * time.Millisecond, // Default
		functionCount:   0,
	}
}

func (m *Manager) SetUnlimitedOutput(unlimited bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.unlimitedOutput = unlimited
}

func (m *Manager) SetUpdateInterval(interval time.Duration) {
	m.displayTick = interval
}

func (m *Manager) Pause() {
	if !m.isPaused {
		m.pauseCh <- true
		m.isPaused = true
	}
}

func (m *Manager) Resume() {
	if m.isPaused {
		m.pauseCh <- false
		m.isPaused = false
	}
}

func (m *Manager) Register(name string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.functionCount++
	m.outputs[fmt.Sprint(m.functionCount)] = &FunctionOutput{
		Name:        name,
		Status:      "pending",
		StreamLines: []string{},
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
		Tables:      make(map[string]*Table),
		Index:       m.functionCount,
	}
	return fmt.Sprint(m.functionCount)
}

func (m *Manager) SetMessage(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Message = message
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) SetStatus(name, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Status = status
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) GetStatus(name string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if info, exists := m.outputs[name]; exists {
		return info.Status
	}
	return "unknown"
}

func (m *Manager) Complete(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		if !m.unlimitedOutput {
			info.StreamLines = []string{}
		}
		if message == "" {
			info.Message = fmt.Sprintf("Completed %s", info.Name)
		} else {
			info.Message = message
		}
		info.Complete = true
		info.Status = "success"
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) ReportError(name string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Complete = true
		info.Status = "error"
		info.Error = err
		info.LastUpdated = time.Now()
		// Add to global error list
		m.errors = append(m.errors, ErrorReport{
			FunctionName: name,
			Error:        err,
			Time:         time.Now(),
		})
	}
}

func (m *Manager) UpdateStreamOutput(name string, output []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		if m.unlimitedOutput { // just append
			info.StreamLines = append(info.StreamLines, output...)
		} else { // enforce size limit
			currentLen := len(info.StreamLines)
			if currentLen+len(output) > m.maxStreams {
				startIndex := currentLen + len(output) - m.maxStreams
				if startIndex > currentLen {
					startIndex = 0
				}
				newLines := append(info.StreamLines[startIndex:], output...)
				if len(newLines) > m.maxStreams {
					newLines = newLines[len(newLines)-m.maxStreams:]
				}
				info.StreamLines = newLines
			} else {
				info.StreamLines = append(info.StreamLines, output...)
			}
		}
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) AddStreamLine(name string, line string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		// Wrap the line with indentation
		wrappedLines := wrapText(line, basePadding+4)
		if m.unlimitedOutput { // just append all wrapped lines
			info.StreamLines = append(info.StreamLines, wrappedLines...)
		} else { // enforce size limit
			currentLen := len(info.StreamLines)
			totalNewLines := len(wrappedLines)
			if currentLen+totalNewLines > m.maxStreams {
				startIndex := currentLen + totalNewLines - m.maxStreams
				if startIndex > currentLen {
					startIndex = 0
					existingToKeep := m.maxStreams - totalNewLines
					if existingToKeep > 0 {
						info.StreamLines = info.StreamLines[currentLen-existingToKeep:]
					} else {
						info.StreamLines = []string{} // All existing lines will be dropped
					}
				} else {
					info.StreamLines = info.StreamLines[startIndex:]
				}
				info.StreamLines = append(info.StreamLines, wrappedLines...)
			} else {
				info.StreamLines = append(info.StreamLines, wrappedLines...)
			}
			if len(info.StreamLines) > m.maxStreams {
				info.StreamLines = info.StreamLines[len(info.StreamLines)-m.maxStreams:]
			}
		}
		info.LastUpdated = time.Now()
	}
}

func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80 // Default fallback width if terminal width can't be determined
	}
	return width
}

func wrapText(text string, indent int) []string {
	termWidth := GetTerminalWidth()
	maxWidth := termWidth - indent - 2 // Account for indentation
	if maxWidth <= 10 {
		maxWidth = 80
	}
	if utf8.RuneCountInString(text) <= maxWidth {
		return []string{text}
	}
	var lines []string
	currentLine := ""
	currentWidth := 0
	for _, r := range text {
		runeWidth := 1
		// If adding this rune would exceed max width, flush the line
		if currentWidth+runeWidth > maxWidth {
			lines = append(lines, currentLine)
			currentLine = string(r)
			currentWidth = runeWidth
		} else {
			currentLine += string(r)
			currentWidth += runeWidth
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

func (m *Manager) AddProgressBarToStream(name string, outof, final int64, text string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		progressBar := PrintProgressBar(max(0, outof), final, 30)
		display := progressBar + debugStyle.Render(text)
		info.StreamLines = []string{display} // Set as only stream so nothing else is displayed
		info.LastUpdated = time.Now()
	}
}

func PrintProgressBar(current, total int64, width int) string {
	if width <= 0 {
		width = 30
	}
	percent := float64(current) / float64(total)
	filled := min(int(percent*float64(width)), width)
	bar := StyleSymbols["bullet"]
	bar += strings.Repeat(StyleSymbols["hline"], filled)
	if filled < width {
		bar += strings.Repeat(" ", width-filled)
	}
	bar += StyleSymbols["bullet"]
	return debugStyle.Render(fmt.Sprintf("%s %.1f%% %s ", bar, percent*100, StyleSymbols["bullet"]))
}

func (m *Manager) ClearLines(n int) {
	if n <= 0 {
		return
	}
	fmt.Printf("\033[%dA\033[J", min(m.numLines, n))
	m.numLines = max(m.numLines-n, 0)
}

func (m *Manager) ClearFunction(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.StreamLines = []string{}
		info.Message = ""
	}
}

func (m *Manager) ClearAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for name := range m.outputs {
		m.outputs[name].StreamLines = []string{}
	}
}

func (m *Manager) GetStatusIndicator(status string) string {
	switch status {
	case "success", "pass":
		return successStyle.Render(StyleSymbols["pass"])
	case "error", "fail":
		return errorStyle.Render(StyleSymbols["fail"])
	case "warning":
		return warningStyle.Render(StyleSymbols["warning"])
	case "pending":
		return pendingStyle.Render(StyleSymbols["pending"])
	default:
		return infoStyle.Render(StyleSymbols["bullet"])
	}
}

// Add a global table
func (m *Manager) RegisterTable(name string, headers []string) *Table {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	table := NewTable(headers)
	m.tables[name] = table
	return table
}

// Adds a function-specific table
func (m *Manager) RegisterFunctionTable(funcName string, name string, headers []string) *Table {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[funcName]; exists {
		table := NewTable(headers)
		info.Tables[name] = table
		return table
	}
	return nil
}

func (m *Manager) sortFunctions() (active, pending, completed []*FunctionOutput) {
	var allFuncs []*FunctionOutput
	// Sort by index (registration order)
	for _, info := range m.outputs {
		allFuncs = append(allFuncs, info)
	}
	sort.Slice(allFuncs, func(i, j int) bool {
		return allFuncs[i].Index < allFuncs[j].Index
	})
	// Group functions by status
	for _, f := range allFuncs {
		if f.Complete {
			completed = append(completed, f)
		} else if f.Status == "pending" && f.Message == "" {
			pending = append(pending, f)
		} else {
			active = append(active, f)
		}
	}
	return active, pending, completed
}

func (m *Manager) updateDisplay() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.numLines > 0 && !m.unlimitedOutput {
		fmt.Printf("\033[%dA\033[J", m.numLines)
	}
	lineCount := 0
	activeFuncs, pendingFuncs, completedFuncs := m.sortFunctions()

	// Display active functions
	for _, f := range activeFuncs {
		info := f
		statusDisplay := m.GetStatusIndicator(info.Status)
		elapsed := time.Since(info.StartTime).Round(time.Second)
		if info.Complete {
			elapsed = info.LastUpdated.Sub(info.StartTime).Round(time.Second)
		}
		elapsedStr := elapsed.String()

		// Style the message based on status
		var styledMessage string
		switch info.Status {
		case "success":
			styledMessage = successStyle.Render(info.Message)
		case "error":
			styledMessage = errorStyle.Render(info.Message)
		case "warning":
			styledMessage = warningStyle.Render(info.Message)
		default: // pending or other
			styledMessage = pendingStyle.Render(info.Message)
		}
		fmt.Printf("%s%s %s %s\n", strings.Repeat(" ", basePadding), statusDisplay, debugStyle.Render(elapsedStr), styledMessage)
		lineCount++

		// Print stream lines with indentation
		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4) // Additional indentation for stream output
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Display pending functions
	for _, f := range pendingFuncs {
		info := f
		statusDisplay := m.GetStatusIndicator(info.Status)
		fmt.Printf("%s%s %s\n", strings.Repeat(" ", basePadding), statusDisplay, pendingStyle.Render("Waiting..."))
		lineCount++
		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4)
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Display completed functions
	for _, f := range completedFuncs {
		info := f
		statusDisplay := m.GetStatusIndicator(info.Status)
		totalTime := info.LastUpdated.Sub(info.StartTime).Round(time.Second)
		timeStr := totalTime.String()

		// Style message based on status
		var styledMessage string
		if info.Status == "success" {
			styledMessage = successStyle.Render(info.Message)
		} else if info.Status == "error" {
			styledMessage = errorStyle.Render(info.Message)
		} else if info.Status == "warning" {
			styledMessage = warningStyle.Render(info.Message)
		} else { // pending or other
			styledMessage = pendingStyle.Render(info.Message)
		}
		fmt.Printf("%s%s %s %s\n", strings.Repeat(" ", basePadding), statusDisplay, debugStyle.Render(timeStr), styledMessage)
		lineCount++

		// Print stream lines with indentation if unlimited mode is enabled
		if m.unlimitedOutput && len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4)
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}
	m.numLines = lineCount
}

func (m *Manager) StartDisplay() {
	m.displayWg.Add(1)
	go func() {
		defer m.displayWg.Done()
		ticker := time.NewTicker(m.displayTick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !m.isPaused {
					m.updateDisplay()
				}
			case pauseState := <-m.pauseCh:
				m.isPaused = pauseState
			case <-m.doneCh:
				if !m.unlimitedOutput {
					m.ClearAll()
				}
				m.updateDisplay()
				m.ShowSummary()
				m.displayTables()
				return
			}
		}
	}()
}

func (m *Manager) StopDisplay() {
	close(m.doneCh)
	m.displayWg.Wait() // Wait for goroutine to finish
}

func (m *Manager) displayTables() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.tables) > 0 {
		fmt.Println(strings.Repeat(" ", basePadding) + headerStyle.Render("Global Tables:"))
		for name, table := range m.tables {
			fmt.Println(strings.Repeat(" ", basePadding+2) + headerStyle.Render(name))
			fmt.Println(table.FormatTable(false))
		}
	}
	// Display function tables
	hasFunctionTables := false
	for _, info := range m.outputs {
		if len(info.Tables) > 0 {
			hasFunctionTables = true
			break
		}
	}
	if hasFunctionTables {
		fmt.Println(strings.Repeat(" ", basePadding) + headerStyle.Render("Function Tables:"))
		for _, info := range m.outputs {
			if len(info.Tables) > 0 {
				fmt.Println(strings.Repeat(" ", basePadding+2) + headerStyle.Render(info.Name))
				for tableName, table := range info.Tables {
					fmt.Println(strings.Repeat(" ", basePadding+4) + infoStyle.Render(tableName))
					fmt.Println(table.FormatTable(false))
				}
			}
		}
	}
}

func (m *Manager) displayErrors() {
	if len(m.errors) == 0 {
		return
	}
	fmt.Println()
	fmt.Println(strings.Repeat(" ", basePadding) + errorStyle.Bold(true).Render("Errors:"))
	for i, err := range m.errors {
		fmt.Printf("%s%s %s %s\n",
			strings.Repeat(" ", basePadding+2),
			errorStyle.Render(fmt.Sprintf("%d.", i+1)),
			debugStyle.Render(fmt.Sprintf("[%s]", err.Time.Format("15:04:05"))),
			errorStyle.Render(fmt.Sprintf("Function: %s", err.FunctionName)))
		fmt.Printf("%s%s\n", strings.Repeat(" ", basePadding+4), errorStyle.Render(fmt.Sprintf("Error: %v", err.Error)))
	}
}

func (m *Manager) ShowSummary() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	fmt.Println()
	var success, failures int
	for _, info := range m.outputs {
		if info.Status == "success" {
			success++
		} else if info.Status == "error" {
			failures++
		}
	}
	// totalOps := fmt.Sprintf("Total Operations: %d,", len(m.outputs))
	succeeded := fmt.Sprintf("Completed %d of %d", success, len(m.outputs))
	failed := fmt.Sprintf("Failed %d of %d", failures, len(m.outputs))
	fmt.Println(strings.Repeat(" ", basePadding) + success2Style.Render(succeeded))
	if failures > 0 {
		fmt.Println(strings.Repeat(" ", basePadding) + errorStyle.Render(failed))
	}
	m.displayErrors()
	fmt.Println()
}
