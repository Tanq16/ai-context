package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
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

// =========================================== ==============
// =========================================== Output Manager
// =========================================== ==============

// Output manager main structure
type Manager struct {
	status      string
	message     string
	progress    string
	complete    bool
	startTime   time.Time
	lastUpdated time.Time
	mutex       sync.RWMutex
	err         error
	doneCh      chan struct{}  // Channel to signal stopping the display
	displayTick time.Duration  // Interval between display updates
	displayWg   sync.WaitGroup // WaitGroup for display goroutine shutdown
}

func NewManager() *Manager {
	return &Manager{
		status:      "pending",
		message:     "",
		progress:    "",
		complete:    false,
		startTime:   time.Now(),
		lastUpdated: time.Now(),
		err:         nil,
		doneCh:      make(chan struct{}),
		displayTick: 200 * time.Millisecond, // Default
	}
}

func (m *Manager) SetUpdateInterval(interval time.Duration) {
	m.displayTick = interval
}

func (m *Manager) SetMessage(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.message = message
	m.lastUpdated = time.Now()
}

func (m *Manager) SetStatus(status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.status = status
	m.lastUpdated = time.Now()
}

func (m *Manager) GetStatus() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.status
}

func (m *Manager) Complete(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.progress = ""
	if message == "" {
		m.message = "Completed"
	} else {
		m.message = message
	}
	m.complete = true
	m.status = "success"
	m.lastUpdated = time.Now()
}

func (m *Manager) ReportError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.complete = true
	m.status = "error"
	m.err = err
	m.lastUpdated = time.Now()
	m.err = err
}

func (m *Manager) AddProgressBarToStream(outof, final int64, text string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	progressBar := PrintProgressBar(max(0, outof), final, 30)
	display := progressBar + debugStyle.Render(text)
	m.progress = display
	m.lastUpdated = time.Now()
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

func (m *Manager) ClearAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.progress = ""
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

func (m *Manager) updateDisplay() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	fmt.Printf("\033[2A\033[J") // clear 2 lines

	if !m.complete && m.status == "pending" && m.message == "" {
		statusDisplay := m.GetStatusIndicator(m.status)
		fmt.Printf("%s%s %s\n", strings.Repeat(" ", basePadding), statusDisplay, pendingStyle.Render("Waiting..."))
		indent := strings.Repeat(" ", basePadding+4)
		fmt.Printf("%s%s\n", indent, streamStyle.Render(m.progress))
	} else if m.complete {
		statusDisplay := m.GetStatusIndicator(m.status)
		totalTime := m.lastUpdated.Sub(m.startTime).Round(time.Second)
		timeStr := totalTime.String()

		// Style message based on status
		var styledMessage string
		switch m.status {
		case "success":
			styledMessage = successStyle.Render(m.message)
		case "error":
			styledMessage = errorStyle.Render(m.message)
		case "warning":
			styledMessage = warningStyle.Render(m.message)
		default: // pending or other
			styledMessage = pendingStyle.Render(m.message)
		}
		fmt.Printf("%s%s %s %s\n\n", strings.Repeat(" ", basePadding), statusDisplay, debugStyle.Render(timeStr), styledMessage)
	} else {
		statusDisplay := m.GetStatusIndicator(m.status)
		elapsed := time.Since(m.startTime).Round(time.Second)
		if m.complete {
			elapsed = m.lastUpdated.Sub(m.startTime).Round(time.Second)
		}
		elapsedStr := elapsed.String()

		// Style the message based on status
		var styledMessage string
		switch m.status {
		case "success":
			styledMessage = successStyle.Render(m.message)
		case "error":
			styledMessage = errorStyle.Render(m.message)
		case "warning":
			styledMessage = warningStyle.Render(m.message)
		default: // pending or other
			styledMessage = pendingStyle.Render(m.message)
		}
		fmt.Printf("%s%s %s %s\n", strings.Repeat(" ", basePadding), statusDisplay, debugStyle.Render(elapsedStr), styledMessage)

		// Print stream lines with indentation
		indent := strings.Repeat(" ", basePadding+4) // Additional indentation for stream output
		fmt.Printf("%s%s\n", indent, streamStyle.Render(m.progress))
	}
}

func (m *Manager) StartDisplay() {
	fmt.Println()
	m.displayWg.Add(1)
	go func() {
		defer m.displayWg.Done()
		ticker := time.NewTicker(m.displayTick)
		defer ticker.Stop()
		fmt.Print("\n\n")
		m.updateDisplay() // update at start and print 2 lines for compensation
		for {
			select {
			case <-ticker.C:
				m.updateDisplay()
			case <-m.doneCh:
				m.ClearAll()
				m.updateDisplay()
				m.ShowSummary()
				return
			}
		}
	}()
}

func (m *Manager) StopDisplay() {
	close(m.doneCh)
	m.displayWg.Wait() // Wait for goroutine to finish
}

func (m *Manager) displayError() {
	if m.err == nil {
		return
	}
	fmt.Println()
	fmt.Println(strings.Repeat(" ", basePadding) + errorStyle.Bold(true).Render("Encountered Error:"))
	fmt.Printf("%s%s\n",
		strings.Repeat(" ", basePadding+2),
		debugStyle.Render(fmt.Sprintf("%s", m.err)),
	)
}

func (m *Manager) ShowSummary() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.displayError()
}
