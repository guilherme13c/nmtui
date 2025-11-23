package main

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	networks       []string
	activeNetworks []string
	cursor         int
}

func NewModel() *Model {
	return &Model{
		networks:       []string{},
		activeNetworks: []string{},
		cursor:         0,
	}
}

func (m *Model) Init() tea.Cmd {
	cmd := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		if !slices.Contains([]string{"", "lo"}, line) {
			m.networks = append(m.networks, line)
		}
	}

	cmd = exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show", "--active")
	output, err = cmd.Output()
	if err != nil {
		return nil
	}

	m.activeNetworks = strings.Split(string(output), "\n")

	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.networks)-1 {
				m.cursor++
			}
		case "enter", " ", "c":
			selected := m.networks[m.cursor]
			cmd := exec.Command("nmcli", "connection", "up", selected)
			cmd.Run()
		case "d", "backspace":
			selected := m.networks[m.cursor]
			cmd := exec.Command("nmcli", "connection", "down", selected)
			cmd.Run()
		}

	}
	return m, nil
}

func (m *Model) View() string {
	s := ""

	for i, network := range m.networks {
		cursor := " "
		if m.cursor == i {
			cursor = "->"
		}

		displayNetwork := network
		if slices.Contains(m.activeNetworks, network) {
			displayNetwork = lipgloss.NewStyle().Background(lipgloss.Color("2")).Render(network)
		}

		s += fmt.Sprintf("%s\t%s\n", cursor, displayNetwork)
	}

	return s
}
