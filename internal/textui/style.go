package textui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
