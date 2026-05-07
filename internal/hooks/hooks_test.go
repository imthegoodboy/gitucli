package hooks

import (
	"strings"
	"testing"
)

func TestScriptIncludesDBFlagWhenProvided(t *testing.T) {
	body := script("pre-commit", `C:\tools\gitu.exe`, `C:\tmp\gitu.db`)
	if !strings.Contains(body, `--db "C:/tmp/gitu.db" guard pre-commit`) {
		t.Fatalf("script does not preserve db path:\n%s", body)
	}
}

func TestScriptOmitsDBFlagWhenEmpty(t *testing.T) {
	body := script("pre-push", `C:\tools\gitu.exe`, "")
	if strings.Contains(body, "--db") {
		t.Fatalf("script should not include db flag:\n%s", body)
	}
}
