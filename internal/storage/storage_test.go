package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestProfilesAndRepoMappings(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "gitu.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	profile, err := store.SaveProfile(ctx, Profile{
		Name:       "work",
		GitHubUser: "work-user",
		GitName:    "Work User",
		Email:      "work@example.com",
		SSHKeyPath: filepath.Join(t.TempDir(), "key"),
		SSHAlias:   "github-work",
	})
	if err != nil {
		t.Fatal(err)
	}

	mapping, err := store.SaveRepoMapping(ctx, RepoMapping{
		RepoPath:       filepath.Join(t.TempDir(), "repo"),
		ProfileID:      profile.ID,
		RemoteName:     "origin",
		OriginalRemote: "git@github.com:owner/repo.git",
		ExpectedRemote: "git@github-work:owner/repo.git",
		HookMode:       "block",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.GetRepoMapping(ctx, mapping.RepoPath)
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpectedRemote != "git@github-work:owner/repo.git" {
		t.Fatalf("unexpected expected remote: %s", got.ExpectedRemote)
	}
}
