package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/parth/gitucli/internal/storage"
)

func TestPersistentDBFlagAppliesToSubcommands(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "custom.db")
	sshConfigPath := filepath.Join(dir, "ssh", "config")
	keyPath := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITU_SSH_CONFIG", sshConfigPath)

	var out, errOut bytes.Buffer
	root := newRootCommand(&commandEnv{
		in:     bytes.NewBuffer(nil),
		out:    &out,
		errOut: &errOut,
	})
	root.SetArgs([]string{
		"--db", dbPath,
		"profile", "add",
		"--name", "work",
		"--github-user", "work-user",
		"--git-name", "Work User",
		"--email", "work@example.com",
		"--key", keyPath,
		"--alias", "github-work",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("command failed: %v\nstderr: %s", err, errOut.String())
	}

	store, err := storage.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	profile, err := store.GetProfileByName(context.Background(), "work")
	if err != nil {
		t.Fatal(err)
	}
	if profile.SSHAlias != "github-work" {
		t.Fatalf("profile was not saved in custom DB: %#v", profile)
	}
}

func TestRootCommandRunsInteractiveShellAndExits(t *testing.T) {
	var out, errOut bytes.Buffer
	root := newRootCommand(&commandEnv{
		in:     bytes.NewBufferString("0\n"),
		out:    &out,
		errOut: &errOut,
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("interactive root failed: %v\nstderr: %s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "Setup a repo") {
		t.Fatalf("interactive menu did not render expected options:\n%s", out.String())
	}
}
