package textui

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd { return tea.Quit }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}
func (m model) View() string {
	title := Title("gituCli")
	subtitle := Muted("Multi GitHub Identity Manager")
	menu := []string{
		Command("gitu init [path]") + "          Initialize or attach a repo",
		Command("gitu profile add") + "          Add a GitHub identity profile",
		Command("gitu validate [path]") + "      Check repo identity safety",
		Command("gitu repair [path]") + "        Restore managed repo settings",
		Command("gitu autocommit [path]") + "    Commit changes safely on demand or schedule",
		Command("gitu key generate <name>") + "  Generate an SSH key",
		Command("gitu daemon") + "               Watch configured repos",
	}
	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s\n",
		title,
		subtitle,
		strings.Join(menu, "\n"),
		Muted("Tip: run gitu init --help for non-interactive flags."),
	)
}

func Render(w io.Writer) error {
	_, err := fmt.Fprint(w, model{}.View())
	return err
}
