package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/parth/gitucli/internal/gitutil"
)

type AutoCommitOptions struct {
	RepoPath   string
	Message    string
	Push       bool
	RemoteName string
	DryRun     bool
}

type AutoCommitResult struct {
	RepoPath       string
	Committed      bool
	Pushed         bool
	Skipped        bool
	CommitMessage  string
	StatusBefore   []string
	Validation     Report
	ValidationOnly bool
}

func (s *Service) AutoCommitOnce(ctx context.Context, opts AutoCommitOptions) (AutoCommitResult, error) {
	message := strings.TrimSpace(opts.Message)
	if message == "" {
		message = defaultAutoCommitMessage()
	}

	report, err := s.ValidateRepo(ctx, opts.RepoPath)
	if err != nil {
		return AutoCommitResult{}, err
	}
	result := AutoCommitResult{
		RepoPath:      report.RepoPath,
		CommitMessage: message,
		Validation:    report,
	}
	if !report.OK {
		return result, fmt.Errorf("repo identity is unsafe; run gitu repair %q", report.RepoPath)
	}

	status, err := gitutil.PorcelainStatus(ctx, report.RepoPath)
	if err != nil {
		return result, err
	}
	result.StatusBefore = gitutil.ShortStatusLines(status)
	if strings.TrimSpace(status) == "" {
		result.Skipped = true
		return result, nil
	}
	if opts.DryRun {
		result.ValidationOnly = true
		return result, nil
	}

	if err := gitutil.AddAll(ctx, report.RepoPath); err != nil {
		return result, err
	}
	if err := gitutil.Commit(ctx, report.RepoPath, message); err != nil {
		return result, err
	}
	result.Committed = true

	if opts.Push {
		if err := gitutil.Push(ctx, report.RepoPath, opts.RemoteName); err != nil {
			return result, err
		}
		result.Pushed = true
	}
	return result, nil
}

func defaultAutoCommitMessage() string {
	return "auto commit " + time.Now().Format("2006-01-02 15:04:05")
}

func NextClockDelay(now time.Time, clock string) (time.Duration, error) {
	clock = strings.TrimSpace(clock)
	if clock == "" {
		return 0, nil
	}
	parsed, err := time.Parse("15:04", clock)
	if err != nil {
		return 0, fmt.Errorf("time must use HH:MM format, got %q", clock)
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now), nil
}
