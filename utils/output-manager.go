package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))  // dark green
	success2Style = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // green
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))   // red
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))  // yellow
	pendingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // blue
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))  // cyan
	debugStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // light grey
	detailStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))  // purple
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

// =========================================== ==============
// =========================================== Output Manager
// =========================================== ==============

// Output manager main structure
type Manager struct {
	status    string
	message   string
	progress  string
	complete  bool
	startTime time.Time
	mutex     sync.RWMutex
	err       error
	doneCh    chan struct{}  // Channel to signal stopping the display
	displayWg sync.WaitGroup // WaitGroup for display goroutine shutdown
	disabled  bool
}

func NewManager() *Manager {
	return &Manager{
		status:    "pending",
		startTime: time.Now(),
		doneCh:    make(chan struct{}),
		disabled:  false,
	}
}

func (m *Manager) Disable() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.disabled = true
}

func (m *Manager) SetMessage(message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.message = message
}

func (m *Manager) Complete(message string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.progress = ""
	if message == "" {
		m.message = "Completed"
	} else {
		m.message = message
	}
	m.complete = true
	if err != nil {
		m.status = "error"
		m.err = err
	} else {
		m.status = "success"
	}
}

func (m *Manager) ReportProgress(outof, final int64, text string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	progressBar := printProgressBar(max(0, outof), final, 30)
	display := progressBar + debugStyle.Render(text)
	m.progress = display
}

func printProgressBar(current, total int64, width int) string {
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

func (m *Manager) StopDisplay() {
	close(m.doneCh)
	m.displayWg.Wait() // Wait for goroutine to finish
}

func (m *Manager) getStatusIndicator() string {
	switch m.status {
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

func (m *Manager) getStyledMessage() string {
	var styledMessage string
	switch m.status {
	case "success":
		styledMessage = successStyle.Render(m.message)
	case "error":
		styledMessage = errorStyle.Render(m.message)
	case "warning":
		styledMessage = warningStyle.Render(m.message)
	default: // pending or unknown
		styledMessage = pendingStyle.Render(m.message)
	}
	return styledMessage
}

func (m *Manager) updateDisplay() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.disabled {
		return
	}
	fmt.Printf("\033[2A\033[J") // clear 2 lines
	if !m.complete && m.status == "pending" && m.message == "" {
		// case of unprocessed jobs
		statusDisplay := m.getStatusIndicator()
		fmt.Printf("  %s %s\n", statusDisplay, pendingStyle.Render("Waiting..."))
		fmt.Printf("      %s\n", debugStyle.Render(m.progress))
	} else {
		statusDisplay := m.getStatusIndicator()
		totalTime := time.Since(m.startTime).Round(time.Second)
		elapsedStr := totalTime.String()
		fmt.Printf("  %s %s %s\n", statusDisplay, debugStyle.Render(elapsedStr), m.getStyledMessage())
		fmt.Printf("      %s\n", debugStyle.Render(m.progress))
	}
}

func (m *Manager) StartDisplay() {
	m.displayWg.Add(1)
	go func() {
		defer m.displayWg.Done()
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		fmt.Print("\n\n\n")
		m.updateDisplay() // update at start and print 2 extra lines for compensation
		for {
			select {
			case <-ticker.C:
				m.updateDisplay()
			case <-m.doneCh:
				m.progress = ""
				m.updateDisplay()
				m.checkAndShowError()
				return
			}
		}
	}()
}

func (m *Manager) checkAndShowError() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.disabled || m.err == nil {
		return
	}
	fmt.Println("  " + errorStyle.Bold(true).Render("Encountered Errors:"))
	fmt.Printf("    %s\n", debugStyle.Render(fmt.Sprintf("%s", m.err)))
}
