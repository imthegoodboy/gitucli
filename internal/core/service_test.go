package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/parth/gitucli/internal/gitutil"
	"github.com/parth/gitucli/internal/storage"
)

func TestConfigureTwoReposWithDifferentProfiles(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "gitu.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	svc, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	svc.SSHConfigPath = filepath.Join(t.TempDir(), ".ssh", "config")
	svc.ExePath = "gitu"

	keyA := touchKey(t, "a")
	keyB := touchKey(t, "b")

	if _, err := svc.SaveProfile(ctx, ProfileInput{
		Name:       "account-a",
		GitHubUser: "accountA",
		GitName:    "Account A",
		Email:      "a@example.com",
		SSHKeyPath: keyA,
		SSHAlias:   "github-account-a",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveProfile(ctx, ProfileInput{
		Name:       "account-b",
		GitHubUser: "accountB",
		GitName:    "Account B",
		Email:      "b@example.com",
		SSHKeyPath: keyB,
		SSHAlias:   "github-account-b",
	}); err != nil {
		t.Fatal(err)
	}

	repoA := filepath.Join(t.TempDir(), "repo-a")
	repoB := filepath.Join(t.TempDir(), "repo-b")

	reportA, err := svc.ConfigureRepo(ctx, InitOptions{
		RepoPath:    repoA,
		ProfileName: "account-a",
		RemoteName:  "origin",
		RepoSlug:    "owner/repo-a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reportA.OK {
		t.Fatalf("repo A not OK: %#v", reportA.Issues)
	}

	reportB, err := svc.ConfigureRepo(ctx, InitOptions{
		RepoPath:    repoB,
		ProfileName: "account-b",
		RemoteName:  "origin",
		RepoSlug:    "owner/repo-b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reportB.OK {
		t.Fatalf("repo B not OK: %#v", reportB.Issues)
	}

	emailA, err := gitutil.GetLocalConfig(ctx, repoA, "user.email")
	if err != nil {
		t.Fatal(err)
	}
	emailB, err := gitutil.GetLocalConfig(ctx, repoB, "user.email")
	if err != nil {
		t.Fatal(err)
	}
	if emailA != "a@example.com" || emailB != "b@example.com" {
		t.Fatalf("emails crossed: repoA=%s repoB=%s", emailA, emailB)
	}

	remoteA, err := gitutil.RemoteGetURL(ctx, repoA, "origin")
	if err != nil {
		t.Fatal(err)
	}
	remoteB, err := gitutil.RemoteGetURL(ctx, repoB, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if remoteA != "git@github-account-a:owner/repo-a.git" {
		t.Fatalf("unexpected repo A remote: %s", remoteA)
	}
	if remoteB != "git@github-account-b:owner/repo-b.git" {
		t.Fatalf("unexpected repo B remote: %s", remoteB)
	}
}

func touchKey(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("dummy key"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

