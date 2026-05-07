package textui

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd { return tea.Quit }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}
func (m model) View() string {
	return Dashboard([]MenuRow{
		{Index: "1", Command: "gitu init [path]", Description: "Initialize or attach a repo"},
		{Index: "2", Command: "gitu profile add", Description: "Add a GitHub identity profile"},
		{Index: "3", Command: "gitu validate [path]", Description: "Check repo identity safety"},
		{Index: "4", Command: "gitu repair [path]", Description: "Restore managed repo settings"},
		{Index: "5", Command: "gitu autocommit [path]", Description: "Commit safely now or on schedule"},
		{Index: "6", Command: "gitu key generate <name>", Description: "Generate an SSH key"},
		{Index: "7", Command: "gitu daemon", Description: "Watch configured repos"},
	})
}

func Render(w io.Writer) error {
	_, err := fmt.Fprint(w, model{}.View())
	return err
}
