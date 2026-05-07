package gitutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func NormalizePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return filepath.Clean(abs), nil
}

func EnsureRepo(ctx context.Context, path string) (string, error) {
	repoPath, err := NormalizePath(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return "", err
	}
	if IsRepo(ctx, repoPath) {
		return TopLevel(ctx, repoPath)
	}
	if _, err := Run(ctx, repoPath, "init"); err != nil {
		return "", err
	}
	return TopLevel(ctx, repoPath)
}

func IsRepo(ctx context.Context, path string) bool {
	out, err := Run(ctx, path, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

func TopLevel(ctx context.Context, path string) (string, error) {
	out, err := Run(ctx, path, "rev-parse", "--show-toplevel")
	if err != nil {
		return NormalizePath(path)
	}
	return NormalizePath(strings.TrimSpace(out))
}

func GitDir(ctx context.Context, repoPath string) (string, error) {
	out, err := Run(ctx, repoPath, "rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	gitDir := strings.TrimSpace(out)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoPath, gitDir)
	}
	return filepath.Clean(gitDir), nil
}

func SetLocalConfig(ctx context.Context, repoPath, key, value string) error {
	_, err := Run(ctx, repoPath, "config", "--local", key, value)
	return err
}

func GetLocalConfig(ctx context.Context, repoPath, key string) (string, error) {
	out, err := Run(ctx, repoPath, "config", "--local", "--get", key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func RemoteGetURL(ctx context.Context, repoPath, remoteName string) (string, error) {
	out, err := Run(ctx, repoPath, "remote", "get-url", remoteName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func RemoteSetURL(ctx context.Context, repoPath, remoteName, remoteURL string) error {
	if _, err := RemoteGetURL(ctx, repoPath, remoteName); err != nil {
		_, addErr := Run(ctx, repoPath, "remote", "add", remoteName, remoteURL)
		return addErr
	}
	_, err := Run(ctx, repoPath, "remote", "set-url", remoteName, remoteURL)
	return err
}

func Run(ctx context.Context, repoPath string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

func IsMissingConfig(err error) bool {
	return err != nil && strings.Contains(err.Error(), "exit status 1")
}

func IsGitUnavailable(err error) bool {
	var execErr *exec.Error
	return errors.As(err, &execErr)
}

