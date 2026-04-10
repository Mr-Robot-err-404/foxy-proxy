package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var serverLabel = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7EC8E3")).
	Bold(true)

type model struct {
	port  string
	width int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	text := serverLabel.Render(fmt.Sprintf("server running on port %s", m.port))
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, text)
}

func runTUI(port string) error {
	m := model{port: port}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
