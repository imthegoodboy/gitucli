package hooks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/parth/gitucli/internal/gitutil"
)

const marker = "# gituCli managed hook"

var Events = []string{"pre-commit", "pre-push"}

func Install(ctx context.Context, repoPath, exePath string) error {
	gitDir, err := gitutil.GitDir(ctx, repoPath)
	if err != nil {
		return err
	}
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}

	for _, event := range Events {
		path := filepath.Join(hooksDir, event)
		if err := writeHook(path, event, exePath); err != nil {
			return err
		}
	}
	return nil
}

func IsInstalled(ctx context.Context, repoPath, event string) bool {
	gitDir, err := gitutil.GitDir(ctx, repoPath)
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(gitDir, "hooks", event))
	return err == nil && strings.Contains(string(data), marker)
}

func writeHook(path, event, exePath string) error {
	if existing, err := os.ReadFile(path); err == nil && !strings.Contains(string(existing), marker) {
		backup := fmt.Sprintf("%s.gitu-backup.%d", path, time.Now().Unix())
		if err := os.Rename(path, backup); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(script(event, exePath)), 0o755)
}

func script(event, exePath string) string {
	exePath = filepath.ToSlash(exePath)
	return fmt.Sprintf(`#!/bin/sh
%s
repo="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
exec "%s" guard %s --repo "$repo"
`, marker, exePath, event)
}

