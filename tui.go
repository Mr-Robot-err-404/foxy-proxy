package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tnBg      = "#1a1b26"
	tnBorder  = "#3b4261"
	tnFg      = "#c0caf5"
	tnComment = "#565f89"
	tnCyan    = "#7dcfff"
	tnBlue    = "#7aa2f7"
	tnGreen   = "#9ece6a"
	tnOrange  = "#ff9e64"
	tnRed     = "#f7768e"
)

var (
	panelStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(tnBg))

	logPanelStyle = panelStyle.
			BorderRight(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderRightForeground(lipgloss.Color(tnBorder)).
			BorderBackground(lipgloss.Color(tnBg))

	logLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tnFg)).
			Background(lipgloss.Color(tnBg))

	logTimestampStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(tnComment)).
				Background(lipgloss.Color(tnBg))

	logLevelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tnGreen)).
			Background(lipgloss.Color(tnBg)).
			Bold(true)

	portLabelStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			PaddingTop(1).
			Foreground(lipgloss.Color(tnComment)).
			Background(lipgloss.Color(tnBg))

	portValueStyle = lipgloss.NewStyle().
			PaddingTop(1).
			Foreground(lipgloss.Color(tnOrange)).
			Background(lipgloss.Color(tnBg)).
			Bold(true)

	statusDotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tnGreen)).
			Background(lipgloss.Color(tnBg))
)

type tickMsg time.Time

type model struct {
	port    string
	logPath string
	vp      viewport.Model
	width   int
	height  int
	ready   bool
	offset  int64
	content string
}

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *model) readNewLines() string {
	f, err := os.Open(m.logPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	f.Seek(m.offset, 0)
	buf := make([]byte, 32*1024)
	n, _ := f.Read(buf)
	if n == 0 {
		return ""
	}
	m.offset += int64(n)
	return string(buf[:n])
}

func colorLogLine(line string) string {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return logLineStyle.Render(line)
	}
	ts := logTimestampStyle.Render(parts[0] + " " + parts[1])
	msg := parts[2]

	var msgRendered string
	switch {
	case strings.Contains(msg, "ERROR") || strings.Contains(msg, "Fatal"):
		msgRendered = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tnRed)).
			Background(lipgloss.Color(tnBg)).
			Render(msg)
	case strings.Contains(msg, "WARN"):
		msgRendered = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tnOrange)).
			Background(lipgloss.Color(tnBg)).
			Render(msg)
	case strings.Contains(msg, "listening") || strings.Contains(msg, "running"):
		msgRendered = logLevelStyle.Render(msg)
	case strings.Contains(msg, "POST") || strings.Contains(msg, "GET"):
		msgRendered = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tnBlue)).
			Background(lipgloss.Color(tnBg)).
			Render(msg)
	default:
		msgRendered = logLineStyle.Render(msg)
	}
	return ts + " " + msgRendered
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.vp = viewport.New(m.logPanelInnerWidth(), m.logPanelHeight())
			m.vp.SetContent(m.content)
			m.ready = true
		} else {
			m.vp.Width = m.logPanelInnerWidth()
			m.vp.Height = m.logPanelHeight()
		}

	case tickMsg:
		chunk := m.readNewLines()
		if chunk != "" {
			lines := strings.Split(strings.TrimRight(chunk, "\n"), "\n")
			colored := make([]string, len(lines))
			for i, l := range lines {
				colored[i] = colorLogLine(l)
			}
			if m.content == "" {
				m.content = strings.Join(colored, "\n")
			} else {
				m.content = m.content + "\n" + strings.Join(colored, "\n")
			}
			m.vp.SetContent(m.content)
			m.vp.GotoBottom()
		}
		return m, tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m model) logPanelWidth() int {
	return m.width / 2
}

func (m model) logPanelInnerWidth() int {
	return m.logPanelWidth() - 1
}

func (m model) logPanelHeight() int {
	return m.height
}

func (m model) rightPanelWidth() int {
	return m.width - m.logPanelWidth()
}

func (m model) View() string {
	if !m.ready {
		return ""
	}

	leftPanel := logPanelStyle.
		Width(m.logPanelWidth()).
		Height(m.height).
		Render(m.vp.View())

	statusLine := lipgloss.JoinHorizontal(lipgloss.Bottom,
		portLabelStyle.Render("port"),
		portValueStyle.Render(fmt.Sprintf(":%s", m.port)),
	)
	rightPanel := panelStyle.
		Width(m.rightPanelWidth()).
		Height(m.height).
		Render(statusLine)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func run_tui(port string, logPath string) error {
	m := model{port: port, logPath: logPath}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
