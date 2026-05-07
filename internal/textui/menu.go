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
	banner := strings.Join([]string{
		"   ____ ___ _____ _   _ ",
		"  / ___|_ _|_   _| | | |",
		" | |  _ | |  | | | | | |",
		" | |_| || |  | | | |_| |",
		"  \\____|___| |_|  \\___/ ",
	}, "\n")
	menu := []string{
		row("1", "gitu init [path]", "Initialize or attach a repo"),
		row("2", "gitu profile add", "Add a GitHub identity profile"),
		row("3", "gitu validate [path]", "Check repo identity safety"),
		row("4", "gitu repair [path]", "Restore managed repo settings"),
		row("5", "gitu autocommit [path]", "Commit safely now or on schedule"),
		row("6", "gitu key generate <name>", "Generate an SSH key"),
		row("7", "gitu daemon", "Watch configured repos"),
	}
	return fmt.Sprintf("%s\n%s\n%s\n\n%s\n\n%s\n%s\n",
		Title(banner),
		Muted("Multi GitHub Identity Manager"),
		Muted("one repo = one GitHub identity"),
		strings.Join(menu, "\n"),
		KeyValue("Storage", "SQLite metadata, no passwords or tokens"),
		KeyValue("Tip", "run "+Command("gitu init --help")+" for non-interactive flags"),
	)
}

func Render(w io.Writer) error {
	_, err := fmt.Fprint(w, model{}.View())
	return err
}

func row(index, command, description string) string {
	return fmt.Sprintf("%s %s  %s", Muted("["+index+"]"), Command(command), description)
}
