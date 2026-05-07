package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/parth/gitucli/internal/storage"
	"github.com/parth/gitucli/internal/textui"
)

func formatCLIError(err error) string {
	if err == nil {
		return ""
	}

	message := cleanError(err.Error())
	suggestion := suggestionFor(err, message)
	return textui.ErrorPanel("gitu failed", message, suggestion)
}

func cleanError(message string) string {
	message = strings.TrimSpace(message)
	message = strings.TrimPrefix(message, "Error: ")
	return message
}

func suggestionFor(err error, message string) string {
	lower := strings.ToLower(message)
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return "Check the name/path, or create the missing profile with: gitu profile add"
	case strings.Contains(lower, "repo is not identity-safe"):
		return "Run gitu validate first, then gitu repair for repairable issues."
	case strings.Contains(lower, "repo is not initialized by gitu"):
		return "Run gitu init <repo-path> --profile <name> --repo owner/name"
	case strings.Contains(lower, "ssh key"):
		return "Generate the key with gitu key generate <profile>, then add the .pub key to the matching GitHub account."
	case strings.Contains(lower, "git email"):
		return "Use an email verified on the intended GitHub account."
	case strings.Contains(lower, "remote"):
		return "Check the GitHub repo slug and run gitu repair if this repo is already managed."
	case strings.Contains(lower, "profile name is required"):
		return "Pass --profile <name> or answer the profile prompt."
	case strings.Contains(lower, "time must use hh:mm"):
		return "Use 24-hour local time, for example: --at 22:30"
	case strings.HasPrefix(lower, "git "):
		return "Run the shown Git command manually for more detail, then rerun gitu validate."
	default:
		return fmt.Sprintf("Run %s for command options.", textui.Command("gitu --help"))
	}
}
