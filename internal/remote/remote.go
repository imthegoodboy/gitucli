package remote

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var scpLike = regexp.MustCompile(`^git@([^:]+):([^/]+)/(.+)$`)

type GitHubRemote struct {
	Host  string
	Owner string
	Repo  string
}

func ParseGitHubRemote(raw string) (GitHubRemote, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return GitHubRemote{}, fmt.Errorf("remote URL is empty")
	}

	if match := scpLike.FindStringSubmatch(raw); match != nil {
		return normalize(match[1], match[2], match[3])
	}

	u, err := url.Parse(raw)
	if err != nil {
		return GitHubRemote{}, fmt.Errorf("parse remote URL: %w", err)
	}

	switch u.Scheme {
	case "ssh":
		if u.User.Username() != "" && u.User.Username() != "git" {
			return GitHubRemote{}, fmt.Errorf("unsupported SSH user %q", u.User.Username())
		}
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) < 2 {
			return GitHubRemote{}, fmt.Errorf("remote URL does not include owner/repo")
		}
		return normalize(u.Hostname(), parts[0], strings.Join(parts[1:], "/"))
	case "https", "http":
		if !strings.EqualFold(u.Hostname(), "github.com") {
			return GitHubRemote{}, fmt.Errorf("only github.com HTTPS remotes are supported")
		}
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) < 2 {
			return GitHubRemote{}, fmt.Errorf("remote URL does not include owner/repo")
		}
		return normalize(u.Hostname(), parts[0], strings.Join(parts[1:], "/"))
	default:
		return GitHubRemote{}, fmt.Errorf("unsupported remote URL format %q", raw)
	}
}

func ParseSlug(slug string) (GitHubRemote, error) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(slug), "/"), "/")
	if len(parts) != 2 {
		return GitHubRemote{}, fmt.Errorf("repo must be in owner/name format")
	}
	return normalize("github.com", parts[0], parts[1])
}

func RewriteToAlias(raw, alias string) (string, error) {
	parsed, err := ParseGitHubRemote(raw)
	if err != nil {
		return "", err
	}
	return BuildAliasRemote(alias, parsed.Owner, parsed.Repo), nil
}

func BuildAliasRemote(alias, owner, repo string) string {
	repo = strings.TrimSuffix(strings.TrimSpace(repo), ".git")
	return fmt.Sprintf("git@%s:%s/%s.git", alias, owner, repo)
}

func (r GitHubRemote) Slug() string {
	return r.Owner + "/" + r.Repo
}

func normalize(host, owner, repo string) (GitHubRemote, error) {
	host = strings.TrimSpace(host)
	owner = strings.Trim(strings.TrimSpace(owner), "/")
	repo = strings.Trim(strings.TrimSpace(repo), "/")
	repo = strings.TrimSuffix(repo, ".git")

	if host == "" || owner == "" || repo == "" {
		return GitHubRemote{}, fmt.Errorf("remote URL is missing host, owner, or repo")
	}
	if strings.Contains(owner, " ") || strings.Contains(repo, " ") {
		return GitHubRemote{}, fmt.Errorf("remote owner/repo cannot contain spaces")
	}

	return GitHubRemote{Host: host, Owner: owner, Repo: repo}, nil
}

