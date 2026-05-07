package cli

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormatCLIErrorIncludesSuggestion(t *testing.T) {
	got := formatCLIError(fmt.Errorf("git email %q is invalid", "bad"))
	if !strings.Contains(got, "gitu failed") {
		t.Fatalf("missing title: %s", got)
	}
	if !strings.Contains(got, "Use an email verified") {
		t.Fatalf("missing useful suggestion: %s", got)
	}
}

