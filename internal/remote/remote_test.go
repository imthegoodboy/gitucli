package remote

import "testing"

func TestParseAndRewriteGitHubRemote(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "scp", input: "git@github.com:owner/repo.git"},
		{name: "ssh", input: "ssh://git@github.com/owner/repo.git"},
		{name: "https", input: "https://github.com/owner/repo.git"},
		{name: "alias", input: "git@github-work:owner/repo.git"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseGitHubRemote(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Owner != "owner" || parsed.Repo != "repo" {
				t.Fatalf("unexpected parse: %#v", parsed)
			}
			rewritten, err := RewriteToAlias(tt.input, "github-work")
			if err != nil {
				t.Fatal(err)
			}
			if rewritten != "git@github-work:owner/repo.git" {
				t.Fatalf("unexpected rewrite: %s", rewritten)
			}
		})
	}
}

func TestParseSlug(t *testing.T) {
	parsed, err := ParseSlug("owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Slug() != "owner/repo" {
		t.Fatalf("unexpected slug: %s", parsed.Slug())
	}
}

