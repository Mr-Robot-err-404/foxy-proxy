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

var headerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7EC8E3")).
	Bold(true)

var logStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#AAAAAA"))

type tickMsg time.Time
type logMsg string

type model struct {
	port     string
	logPath  string
	viewport viewport.Model
	width    int
	height   int
	ready    bool
	offset   int64
}

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) readNewLines() (string, int64) {
	f, err := os.Open(m.logPath)
	if err != nil {
		return "", m.offset
	}
	defer f.Close()

	f.Seek(m.offset, 0)
	buf := make([]byte, 32*1024)
	n, _ := f.Read(buf)
	if n == 0 {
		return "", m.offset
	}
	return string(buf[:n]), m.offset + int64(n)
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
		headerHeight := 2
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight)
			m.viewport.SetContent("")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight
		}

	case tickMsg:
		chunk, newOffset := m.readNewLines()
		if chunk != "" {
			m.offset = newOffset
			existing := m.viewport.View()
			// viewport.View() includes padding/wrapping; track content separately
			updated := strings.TrimRight(existing, "\n") + "\n" + logStyle.Render(strings.TrimRight(chunk, "\n"))
			m.viewport.SetContent(updated)
			m.viewport.GotoBottom()
		}
		return m, tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return ""
	}
	header := lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		headerStyle.Render(fmt.Sprintf("foxy — port %s", m.port)),
	)
	return header + "\n" + m.viewport.View()
}

func runTUI(port string, logPath string) error {
	m := model{port: port, logPath: logPath}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
