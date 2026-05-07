package sshconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	beginMarker = "# >>> gituCli managed identities >>>"
	endMarker   = "# <<< gituCli managed identities <<<"
)

type Entry struct {
	Alias       string
	IdentityFile string
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "config"), nil
}

func RenderManagedBlock(entries []Entry) string {
	entries = append([]Entry(nil), entries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Alias < entries[j].Alias
	})

	var b strings.Builder
	b.WriteString(beginMarker + "\n")
	for _, entry := range entries {
		if strings.TrimSpace(entry.Alias) == "" || strings.TrimSpace(entry.IdentityFile) == "" {
			continue
		}
		b.WriteString("Host " + entry.Alias + "\n")
		b.WriteString("    HostName github.com\n")
		b.WriteString("    User git\n")
		b.WriteString("    IdentityFile " + filepath.ToSlash(entry.IdentityFile) + "\n")
		b.WriteString("    IdentitiesOnly yes\n\n")
	}
	b.WriteString(endMarker + "\n")
	return b.String()
}

func Update(path string, entries []Entry) error {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return err
		}
	}
	path = ExpandHome(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	existingBytes, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	existing := string(existingBytes)
	block := RenderManagedBlock(entries)
	next := replaceBlock(existing, block)

	if existing != "" && existing != next {
		backup := path + ".gitu.bak"
		_ = os.WriteFile(backup, []byte(existing), 0o600)
	}
	return os.WriteFile(path, []byte(next), 0o600)
}

func ValidateKey(path string) error {
	path = ExpandHome(path)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not an SSH key", path)
	}
	return nil
}

func GenerateKey(ctx context.Context, keyPath, email string, force bool) error {
	keyPath = ExpandHome(keyPath)
	if _, err := os.Stat(keyPath); err == nil && !force {
		return fmt.Errorf("SSH key already exists at %s", keyPath)
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0o700); err != nil {
		return err
	}

	args := []string{"-t", "ed25519", "-C", email, "-f", keyPath, "-N", ""}
	if force {
		args = append(args, "-q")
	}
	cmd := exec.CommandContext(ctx, "ssh-keygen", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh-keygen: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func ExpandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~"+string(os.PathSeparator)) || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			rest := strings.TrimPrefix(strings.TrimPrefix(path, "~"+string(os.PathSeparator)), "~/")
			if rest == "" {
				return home
			}
			return filepath.Join(home, filepath.FromSlash(rest))
		}
	}
	return filepath.Clean(os.ExpandEnv(path))
}

func replaceBlock(existing, block string) string {
	existing = strings.ReplaceAll(existing, "\r\n", "\n")
	start := strings.Index(existing, beginMarker)
	end := strings.Index(existing, endMarker)

	if start >= 0 && end >= start {
		end += len(endMarker)
		next := strings.TrimRight(existing[:start], "\n") + "\n\n" + strings.TrimRight(block, "\n")
		tail := strings.TrimLeft(existing[end:], "\n")
		if tail != "" {
			next += "\n\n" + tail
		}
		return strings.TrimLeft(next, "\n") + "\n"
	}

	if strings.TrimSpace(existing) == "" {
		return block
	}
	return strings.TrimRight(existing, "\n") + "\n\n" + block
}

