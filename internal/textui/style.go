package textui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type MenuRow struct {
	Index       string
	Command     string
	Description string
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))
	accentStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("45"))
	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))
	warnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))
	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81"))
	commandBoldStyle = commandStyle.Bold(true)
)

func Title(value string) string {
	return titleStyle.Render(value)
}

func Accent(value string) string {
	return accentStyle.Render(value)
}

func Success(value string) string {
	return successStyle.Render(value)
}

func Warn(value string) string {
	return warnStyle.Render(value)
}

func Error(value string) string {
	return errorStyle.Render(value)
}

func Muted(value string) string {
	return mutedStyle.Render(value)
}

func Command(value string) string {
	return commandBoldStyle.Render(value)
}

func Status(ok bool) string {
	if ok {
		return Success("[OK]")
	}
	return Error("[FAIL]")
}

func IssueSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "warning", "warn":
		return Warn("[WARN]")
	case "error":
		return Error("[ERROR]")
	default:
		return Accent("[" + strings.ToUpper(severity) + "]")
	}
}

func KeyValue(key, value string) string {
	return fmt.Sprintf("%s %s", Muted(key+":"), value)
}

func ErrorPanel(title, message, suggestion string) string {
	lines := []string{
		"+------------------------------------------------------------+",
		"| " + Error(title),
		"+------------------------------------------------------------+",
		"| " + message,
	}
	if strings.TrimSpace(suggestion) != "" {
		lines = append(lines, "| "+Muted("Next:")+" "+suggestion)
	}
	lines = append(lines, "+------------------------------------------------------------+")
	return strings.Join(lines, "\n")
}

func Dashboard(rows []MenuRow) string {
	banner := strings.Join([]string{
		"   ____ ___ _____ _   _ ",
		"  / ___|_ _|_   _| | | |",
		" | |  _ | |  | | | | | |",
		" | |_| || |  | | | |_| |",
		"  \\____|___| |_|  \\___/ ",
	}, "\n")
	body := []string{
		Title(banner),
		Muted("Multi GitHub Identity Manager"),
		Muted("one repo = one GitHub identity"),
		"",
		renderRows(rows),
		"",
		KeyValue("Storage", "SQLite metadata, no passwords or tokens"),
		KeyValue("Mode", "guided app shell"),
	}
	return strings.Join(body, "\n") + "\n"
}

func Page(title string, rows []MenuRow) string {
	return Section(title) + "\n\n" + renderRows(rows) + "\n"
}

func Section(title string) string {
	return Title(title)
}

func renderRows(rows []MenuRow) string {
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%s %s %s",
			Muted("["+row.Index+"]"),
			Command(fmt.Sprintf("%-28s", row.Command)),
			row.Description))
	}
	return strings.Join(lines, "\n")
}
