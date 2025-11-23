package main

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor = lipgloss.Color("62")
	activeColor  = lipgloss.Color("42")
	errorColor   = lipgloss.Color("196")

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(primaryColor).
				Padding(0, 1).
				Bold(true)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("241"))

	activeNetworkStyle = lipgloss.NewStyle().
				Foreground(activeColor).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingTop(1)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(primaryColor).
			Bold(true).
			Padding(0, 1)
)

type Model struct {
	networks       []string
	activeNetworks []string
	cursor         int
	width          int
	height         int
	statusMsg      string
	detailsContent string
}

func NewModel() *Model {
	return &Model{
		networks:       []string{},
		activeNetworks: []string{},
		cursor:         0,
		statusMsg:      "Ready",
		detailsContent: "Select a network to view details...",
	}
}

func (m *Model) Init() tea.Cmd {
	return m.refreshNetworks()
}

func (m *Model) refreshNetworks() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show")
		output, err := cmd.Output()
		if err != nil {
			return nil
		}

		cmdActive := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show", "--active")
		outputActive, errActive := cmdActive.Output()
		if errActive != nil {
			return nil
		}

		var netList []string
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line != "" && line != "lo" {
				netList = append(netList, line)
			}
		}

		activeList := strings.Split(strings.TrimSpace(string(outputActive)), "\n")

		return networkDataMsg{networks: netList, active: activeList}
	}
}

type networkDataMsg struct {
	networks []string
	active   []string
}

type detailsMsg string

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case networkDataMsg:
		m.networks = msg.networks
		m.activeNetworks = msg.active

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
		case "enter", " ":
			selected := m.networks[m.cursor]
			m.statusMsg = fmt.Sprintf("Connecting to %s...", selected)
			c := exec.Command("nmcli", "connection", "up", selected)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return m.refreshNetworks()()
			})

		case "d", "backspace":
			selected := m.networks[m.cursor]
			m.statusMsg = fmt.Sprintf("Disconnecting %s...", selected)
			c := exec.Command("nmcli", "connection", "down", selected)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return m.refreshNetworks()()
			})
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	listWidth := int(float64(m.width) * 0.35)
	detailsWidth := m.width - listWidth - 4
	paneHeight := m.height - 4

	var listBuilder strings.Builder
	for i, network := range m.networks {
		isActive := slices.Contains(m.activeNetworks, network)

		icon := "  "
		if isActive {
			icon = "● "
		}

		line := fmt.Sprintf("%s%s", icon, network)

		if m.cursor == i {
			listBuilder.WriteString(selectedItemStyle.Width(listWidth - 2).Render(line))
		} else if isActive {
			listBuilder.WriteString(itemStyle.Render(activeNetworkStyle.Render(line)))
		} else {
			listBuilder.WriteString(itemStyle.Render(line))
		}
		listBuilder.WriteString("\n")
	}

	leftTitle := titleStyle.Width(listWidth).Render("Networks")
	leftContent := lipgloss.NewStyle().
		Width(listWidth).
		Height(paneHeight).
		Render(listBuilder.String())

	leftPane := paneStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, leftTitle, leftContent),
	)

	rightTitle := titleStyle.Width(detailsWidth).Render("Details")
	rightText := lipgloss.NewStyle().
		Padding(1).
		Render(fmt.Sprintf("%s\n\nSTATUS LOG:\n> %s", m.detailsContent, m.statusMsg))

	rightContent := lipgloss.NewStyle().
		Width(detailsWidth).
		Height(paneHeight).
		Render(rightText)

	rightPane := paneStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, rightTitle, rightContent),
	)

	ui := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	help := helpStyle.Render("  ↑/k: up • ↓/j: down • enter: connect • d: disconnect • q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, ui, help)
}
