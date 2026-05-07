package core

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/parth/gitucli/internal/gitutil"
	"github.com/parth/gitucli/internal/hooks"
	"github.com/parth/gitucli/internal/remote"
	"github.com/parth/gitucli/internal/sshconfig"
	"github.com/parth/gitucli/internal/storage"
)

const HookModeBlock = "block"

var aliasSafe = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

type Service struct {
	Store         *storage.Store
	SSHConfigPath string
	ExePath       string
}

type ProfileInput struct {
	Name       string
	GitHubUser string
	GitName    string
	Email      string
	SSHKeyPath string
	SSHAlias   string
}

type InitOptions struct {
	RepoPath    string
	ProfileName string
	RemoteName  string
	RepoSlug    string
}

type Issue struct {
	Code       string
	Severity   string
	Message    string
	Repairable bool
}

type Report struct {
	RepoPath    string
	ProfileName string
	OK          bool
	Issues      []Issue
}

func NewService(store *storage.Store) (*Service, error) {
	sshPath := os.Getenv("GITU_SSH_CONFIG")
	if sshPath == "" {
		var err error
		sshPath, err = sshconfig.DefaultPath()
		if err != nil {
			return nil, err
		}
	}
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return &Service{Store: store, SSHConfigPath: sshPath, ExePath: exePath}, nil
}

func (s *Service) SaveProfile(ctx context.Context, in ProfileInput) (storage.Profile, error) {
	p, err := normalizeProfile(in)
	if err != nil {
		return storage.Profile{}, err
	}
	saved, err := s.Store.SaveProfile(ctx, p)
	if err != nil {
		return storage.Profile{}, err
	}
	if err := s.SyncSSHConfig(ctx); err != nil {
		return storage.Profile{}, err
	}
	return saved, nil
}

func (s *Service) GenerateKey(ctx context.Context, profileName string, force bool) (storage.Profile, error) {
	p, err := s.Store.GetProfileByName(ctx, profileName)
	if err != nil {
		return storage.Profile{}, err
	}
	if err := sshconfig.GenerateKey(ctx, p.SSHKeyPath, p.Email, force); err != nil {
		return storage.Profile{}, err
	}
	return p, nil
}

func (s *Service) ConfigureRepo(ctx context.Context, opts InitOptions) (Report, error) {
	if strings.TrimSpace(opts.ProfileName) == "" {
		return Report{}, fmt.Errorf("profile name is required")
	}
	if opts.RemoteName == "" {
		opts.RemoteName = "origin"
	}

	repoPath, err := gitutil.EnsureRepo(ctx, opts.RepoPath)
	if err != nil {
		return Report{}, err
	}
	profile, err := s.Store.GetProfileByName(ctx, opts.ProfileName)
	if err != nil {
		return Report{}, err
	}

	originalRemote, remoteErr := gitutil.RemoteGetURL(ctx, repoPath, opts.RemoteName)
	var parsed remote.GitHubRemote
	switch {
	case strings.TrimSpace(opts.RepoSlug) != "":
		parsed, err = remote.ParseSlug(opts.RepoSlug)
	case remoteErr == nil:
		parsed, err = remote.ParseGitHubRemote(originalRemote)
	default:
		err = fmt.Errorf("repo has no %q remote; pass --repo owner/name", opts.RemoteName)
	}
	if err != nil {
		return Report{}, err
	}

	expectedRemote := remote.BuildAliasRemote(profile.SSHAlias, parsed.Owner, parsed.Repo)
	if err := gitutil.SetLocalConfig(ctx, repoPath, "user.name", profile.GitName); err != nil {
		return Report{}, err
	}
	if err := gitutil.SetLocalConfig(ctx, repoPath, "user.email", profile.Email); err != nil {
		return Report{}, err
	}
	if err := gitutil.RemoteSetURL(ctx, repoPath, opts.RemoteName, expectedRemote); err != nil {
		return Report{}, err
	}
	if err := hooks.Install(ctx, repoPath, s.ExePath); err != nil {
		return Report{}, err
	}
	if err := s.SyncSSHConfig(ctx); err != nil {
		return Report{}, err
	}

	_, err = s.Store.SaveRepoMapping(ctx, storage.RepoMapping{
		RepoPath:       repoPath,
		ProfileID:      profile.ID,
		RemoteName:     opts.RemoteName,
		OriginalRemote: originalRemote,
		ExpectedRemote: expectedRemote,
		HookMode:       HookModeBlock,
	})
	if err != nil {
		return Report{}, err
	}
	return s.ValidateRepo(ctx, repoPath)
}

func (s *Service) ValidateRepo(ctx context.Context, path string) (Report, error) {
	repoPath, err := gitutil.TopLevel(ctx, path)
	if err != nil {
		repoPath, _ = gitutil.NormalizePath(path)
	}
	report := Report{RepoPath: repoPath, OK: true}

	mapping, err := s.Store.GetRepoMapping(ctx, repoPath)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			report.add("repo_mapping_missing", "error", "repo is not initialized by gitu", true)
			_ = s.record(ctx, report)
			return report, nil
		}
		return report, err
	}

	profile, err := s.Store.GetProfileByID(ctx, mapping.ProfileID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			report.add("profile_missing", "error", "mapped profile no longer exists", true)
			_ = s.record(ctx, report)
			return report, nil
		}
		return report, err
	}
	report.ProfileName = profile.Name

	if got, err := gitutil.GetLocalConfig(ctx, repoPath, "user.name"); err != nil || got != profile.GitName {
		report.add("git_name_mismatch", "error", fmt.Sprintf("local git user.name must be %q", profile.GitName), true)
	}
	if got, err := gitutil.GetLocalConfig(ctx, repoPath, "user.email"); err != nil || got != profile.Email {
		report.add("git_email_mismatch", "error", fmt.Sprintf("local git user.email must be %q", profile.Email), true)
	}

	remoteURL, err := gitutil.RemoteGetURL(ctx, repoPath, mapping.RemoteName)
	if err != nil {
		report.add("remote_missing", "error", fmt.Sprintf("remote %q is missing", mapping.RemoteName), true)
	} else {
		if remoteURL != mapping.ExpectedRemote {
			report.add("remote_url_mismatch", "error", fmt.Sprintf("remote %q must be %s", mapping.RemoteName, mapping.ExpectedRemote), true)
		}
		if parsed, err := remote.ParseGitHubRemote(remoteURL); err != nil || parsed.Host != profile.SSHAlias {
			report.add("remote_alias_mismatch", "error", fmt.Sprintf("remote must use SSH alias %q", profile.SSHAlias), true)
		}
	}

	if err := sshconfig.ValidateKey(profile.SSHKeyPath); err != nil {
		report.add("ssh_key_missing", "error", fmt.Sprintf("SSH key is not readable at %s", profile.SSHKeyPath), false)
	}

	for _, event := range hooks.Events {
		if !hooks.IsInstalled(ctx, repoPath, event) {
			report.add("hook_missing_"+event, "error", fmt.Sprintf("%s hook is missing or unmanaged", event), true)
		}
	}

	_ = s.record(ctx, report)
	return report, nil
}

func (s *Service) RepairRepo(ctx context.Context, path string) (Report, error) {
	repoPath, err := gitutil.TopLevel(ctx, path)
	if err != nil {
		return Report{}, err
	}
	mapping, err := s.Store.GetRepoMapping(ctx, repoPath)
	if err != nil {
		return Report{}, err
	}
	profile, err := s.Store.GetProfileByID(ctx, mapping.ProfileID)
	if err != nil {
		return Report{}, err
	}

	if err := gitutil.SetLocalConfig(ctx, repoPath, "user.name", profile.GitName); err != nil {
		return Report{}, err
	}
	if err := gitutil.SetLocalConfig(ctx, repoPath, "user.email", profile.Email); err != nil {
		return Report{}, err
	}
	if err := gitutil.RemoteSetURL(ctx, repoPath, mapping.RemoteName, mapping.ExpectedRemote); err != nil {
		return Report{}, err
	}
	if err := hooks.Install(ctx, repoPath, s.ExePath); err != nil {
		return Report{}, err
	}
	if err := s.SyncSSHConfig(ctx); err != nil {
		return Report{}, err
	}
	return s.ValidateRepo(ctx, repoPath)
}

func (s *Service) DaemonSweep(ctx context.Context) ([]Report, error) {
	mappings, err := s.Store.ListRepoMappings(ctx)
	if err != nil {
		return nil, err
	}

	reports := make([]Report, 0, len(mappings))
	for _, mapping := range mappings {
		report, err := s.ValidateRepo(ctx, mapping.RepoPath)
		if err != nil {
			reports = append(reports, Report{
				RepoPath: mapping.RepoPath,
				OK:       false,
				Issues: []Issue{{
					Code:     "validation_failed",
					Severity: "error",
					Message:  err.Error(),
				}},
			})
			continue
		}
		for _, issue := range report.Issues {
			if strings.HasPrefix(issue.Code, "hook_missing_") {
				_ = hooks.Install(ctx, mapping.RepoPath, s.ExePath)
				report, _ = s.ValidateRepo(ctx, mapping.RepoPath)
				break
			}
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func (s *Service) SyncSSHConfig(ctx context.Context) error {
	profiles, err := s.Store.ListProfiles(ctx)
	if err != nil {
		return err
	}
	entries := make([]sshconfig.Entry, 0, len(profiles))
	for _, p := range profiles {
		entries = append(entries, sshconfig.Entry{
			Alias:        p.SSHAlias,
			IdentityFile: p.SSHKeyPath,
		})
	}
	return sshconfig.Update(s.SSHConfigPath, entries)
}

func (r *Report) add(code, severity, message string, repairable bool) {
	r.OK = false
	r.Issues = append(r.Issues, Issue{Code: code, Severity: severity, Message: message, Repairable: repairable})
}

func (s *Service) record(ctx context.Context, report Report) error {
	var summary string
	if report.OK {
		summary = "ok"
	} else {
		parts := make([]string, 0, len(report.Issues))
		for _, issue := range report.Issues {
			parts = append(parts, issue.Code)
		}
		summary = strings.Join(parts, ",")
	}
	return s.Store.RecordValidation(ctx, report.RepoPath, report.OK, summary)
}

func normalizeProfile(in ProfileInput) (storage.Profile, error) {
	name := strings.TrimSpace(in.Name)
	githubUser := strings.TrimSpace(in.GitHubUser)
	gitName := strings.TrimSpace(in.GitName)
	email := strings.TrimSpace(in.Email)
	keyPath := strings.TrimSpace(in.SSHKeyPath)
	alias := strings.TrimSpace(in.SSHAlias)

	if name == "" {
		return storage.Profile{}, fmt.Errorf("profile name is required")
	}
	if githubUser == "" {
		githubUser = name
	}
	if gitName == "" {
		gitName = githubUser
	}
	if email == "" {
		return storage.Profile{}, fmt.Errorf("git email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return storage.Profile{}, fmt.Errorf("git email %q is invalid", email)
	}
	if keyPath == "" {
		keyPath = DefaultKeyPath(name)
	}
	if alias == "" {
		alias = DefaultAlias(githubUser)
	}
	if alias == "github.com" || strings.ContainsAny(alias, " \t\r\n:/\\") {
		return storage.Profile{}, fmt.Errorf("SSH alias %q is invalid", alias)
	}

	return storage.Profile{
		Name:       name,
		GitHubUser: githubUser,
		GitName:    gitName,
		Email:      email,
		SSHKeyPath: sshconfig.ExpandHome(keyPath),
		SSHAlias:   alias,
	}, nil
}

func DefaultAlias(seed string) string {
	seed = strings.ToLower(strings.TrimSpace(seed))
	seed = aliasSafe.ReplaceAllString(seed, "-")
	seed = strings.Trim(seed, "-")
	if seed == "" {
		seed = "account"
	}
	return "github-" + seed
}

func DefaultKeyPath(profileName string) string {
	profileName = strings.ToLower(strings.TrimSpace(profileName))
	profileName = aliasSafe.ReplaceAllString(profileName, "_")
	profileName = strings.Trim(profileName, "_")
	if profileName == "" {
		profileName = "account"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".ssh", "gitu_"+profileName)
	}
	return filepath.Join(home, ".ssh", "gitu_"+profileName)
}
