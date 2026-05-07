package textui

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct{}

func (m model) Init() tea.Cmd { return tea.Quit }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}
func (m model) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("gituCli")
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render("Multi GitHub Identity Manager")
	menu := []string{
		"gitu init [path]          Initialize or attach a repo",
		"gitu profile add          Add a GitHub identity profile",
		"gitu validate [path]      Check repo identity safety",
		"gitu repair [path]        Restore managed repo settings",
		"gitu key generate <name>  Generate an SSH key",
		"gitu daemon               Watch configured repos",
	}
	return fmt.Sprintf("%s\n%s\n\n%s\n", title, subtitle, strings.Join(menu, "\n"))
}

func Render(w io.Writer) error {
	program := tea.NewProgram(model{}, tea.WithOutput(w), tea.WithoutRenderer())
	m, err := program.Run()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, m.View())
	return err
}
