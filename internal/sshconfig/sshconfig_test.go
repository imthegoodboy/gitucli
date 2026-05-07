package sshconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdatePreservesUserConfigAndReplacesManagedBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	original := "Host personal\n    HostName example.com\n\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	err := Update(path, []Entry{{Alias: "github-a", IdentityFile: filepath.Join(dir, "a")}})
	if err != nil {
		t.Fatal(err)
	}
	err = Update(path, []Entry{{Alias: "github-b", IdentityFile: filepath.Join(dir, "b")}})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "Host personal") {
		t.Fatalf("user config was not preserved:\n%s", text)
	}
	if strings.Contains(text, "github-a") {
		t.Fatalf("old managed alias was not replaced:\n%s", text)
	}
	if !strings.Contains(text, "github-b") {
		t.Fatalf("new managed alias missing:\n%s", text)
	}
}

