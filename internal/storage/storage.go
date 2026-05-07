package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const appDirName = "gituCli"

var ErrNotFound = errors.New("not found")

type Store struct {
	db   *sql.DB
	path string
}

type Profile struct {
	ID         int64
	Name       string
	GitHubUser string
	GitName    string
	Email      string
	SSHKeyPath string
	SSHAlias   string
	CreatedAt  time.Time
}

type RepoMapping struct {
	ID             int64
	RepoPath       string
	ProfileID      int64
	RemoteName     string
	OriginalRemote string
	ExpectedRemote string
	HookMode       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func DefaultDBPath() (string, error) {
	if home := os.Getenv("GITU_HOME"); home != "" {
		return filepath.Join(home, "gitu.db"), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appDirName, "gitu.db"), nil
}

func OpenDefault(ctx context.Context) (*Store, error) {
	path, err := DefaultDBPath()
	if err != nil {
		return nil, err
	}
	return Open(ctx, path)
}

func Open(ctx context.Context, dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db, path: dbPath}
	if _, err := s.db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.Init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			github_user TEXT NOT NULL,
			git_name TEXT NOT NULL,
			email TEXT NOT NULL,
			ssh_key_path TEXT NOT NULL,
			ssh_alias TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS repo_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo_path TEXT NOT NULL UNIQUE,
			profile_id INTEGER NOT NULL,
			remote_name TEXT NOT NULL,
			original_remote TEXT NOT NULL DEFAULT '',
			expected_remote TEXT NOT NULL,
			hook_mode TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(profile_id) REFERENCES profiles(id) ON DELETE RESTRICT
		)`,
		`CREATE TABLE IF NOT EXISTS validation_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo_path TEXT NOT NULL,
			ok INTEGER NOT NULL,
			summary TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SaveProfile(ctx context.Context, p Profile) (Profile, error) {
	now := time.Now().UTC()
	existing, err := s.GetProfileByName(ctx, p.Name)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Profile{}, err
	}

	if errors.Is(err, ErrNotFound) {
		res, err := s.db.ExecContext(ctx, `INSERT INTO profiles
			(name, github_user, git_name, email, ssh_key_path, ssh_alias, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			p.Name, p.GitHubUser, p.GitName, p.Email, p.SSHKeyPath, p.SSHAlias, now.Format(time.RFC3339Nano))
		if err != nil {
			return Profile{}, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return Profile{}, err
		}
		p.ID = id
		p.CreatedAt = now
		return p, nil
	}

	_, err = s.db.ExecContext(ctx, `UPDATE profiles
		SET github_user = ?, git_name = ?, email = ?, ssh_key_path = ?, ssh_alias = ?
		WHERE id = ?`,
		p.GitHubUser, p.GitName, p.Email, p.SSHKeyPath, p.SSHAlias, existing.ID)
	if err != nil {
		return Profile{}, err
	}
	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt
	return p, nil
}

func (s *Store) GetProfileByName(ctx context.Context, name string) (Profile, error) {
	return scanProfile(s.db.QueryRowContext(ctx, `SELECT id, name, github_user, git_name, email, ssh_key_path, ssh_alias, created_at
		FROM profiles WHERE name = ?`, name))
}

func (s *Store) GetProfileByID(ctx context.Context, id int64) (Profile, error) {
	return scanProfile(s.db.QueryRowContext(ctx, `SELECT id, name, github_user, git_name, email, ssh_key_path, ssh_alias, created_at
		FROM profiles WHERE id = ?`, id))
}

func (s *Store) ListProfiles(ctx context.Context) ([]Profile, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, github_user, git_name, email, ssh_key_path, ssh_alias, created_at
		FROM profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		p, err := scanProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (s *Store) DeleteProfile(ctx context.Context, name string) error {
	p, err := s.GetProfileByName(ctx, name)
	if err != nil {
		return err
	}

	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM repo_mappings WHERE profile_id = ?`, p.ID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("profile %q is used by %d repo mapping(s)", name, count)
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM profiles WHERE id = ?`, p.ID)
	return err
}

func (s *Store) SaveRepoMapping(ctx context.Context, m RepoMapping) (RepoMapping, error) {
	now := time.Now().UTC()
	existing, err := s.GetRepoMapping(ctx, m.RepoPath)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return RepoMapping{}, err
	}

	if errors.Is(err, ErrNotFound) {
		res, err := s.db.ExecContext(ctx, `INSERT INTO repo_mappings
			(repo_path, profile_id, remote_name, original_remote, expected_remote, hook_mode, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			m.RepoPath, m.ProfileID, m.RemoteName, m.OriginalRemote, m.ExpectedRemote, m.HookMode,
			now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
		if err != nil {
			return RepoMapping{}, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return RepoMapping{}, err
		}
		m.ID = id
		m.CreatedAt = now
		m.UpdatedAt = now
		return m, nil
	}

	original := existing.OriginalRemote
	if original == "" {
		original = m.OriginalRemote
	}
	_, err = s.db.ExecContext(ctx, `UPDATE repo_mappings
		SET profile_id = ?, remote_name = ?, original_remote = ?, expected_remote = ?, hook_mode = ?, updated_at = ?
		WHERE id = ?`,
		m.ProfileID, m.RemoteName, original, m.ExpectedRemote, m.HookMode, now.Format(time.RFC3339Nano), existing.ID)
	if err != nil {
		return RepoMapping{}, err
	}
	m.ID = existing.ID
	m.OriginalRemote = original
	m.CreatedAt = existing.CreatedAt
	m.UpdatedAt = now
	return m, nil
}

func (s *Store) GetRepoMapping(ctx context.Context, repoPath string) (RepoMapping, error) {
	return scanRepoMapping(s.db.QueryRowContext(ctx, `SELECT id, repo_path, profile_id, remote_name, original_remote, expected_remote, hook_mode, created_at, updated_at
		FROM repo_mappings WHERE repo_path = ?`, repoPath))
}

func (s *Store) ListRepoMappings(ctx context.Context) ([]RepoMapping, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, repo_path, profile_id, remote_name, original_remote, expected_remote, hook_mode, created_at, updated_at
		FROM repo_mappings ORDER BY repo_path`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []RepoMapping
	for rows.Next() {
		m, err := scanRepoMapping(rows)
		if err != nil {
			return nil, err
		}
		repos = append(repos, m)
	}
	return repos, rows.Err()
}

func (s *Store) RecordValidation(ctx context.Context, repoPath string, ok bool, summary string) error {
	okInt := 0
	if ok {
		okInt = 1
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO validation_events (repo_path, ok, summary, created_at)
		VALUES (?, ?, ?, ?)`, repoPath, okInt, summary, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProfile(row scanner) (Profile, error) {
	var p Profile
	var created string
	if err := row.Scan(&p.ID, &p.Name, &p.GitHubUser, &p.GitName, &p.Email, &p.SSHKeyPath, &p.SSHAlias, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Profile{}, ErrNotFound
		}
		return Profile{}, err
	}
	p.CreatedAt = parseTime(created)
	return p, nil
}

func scanRepoMapping(row scanner) (RepoMapping, error) {
	var m RepoMapping
	var created, updated string
	if err := row.Scan(&m.ID, &m.RepoPath, &m.ProfileID, &m.RemoteName, &m.OriginalRemote, &m.ExpectedRemote, &m.HookMode, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RepoMapping{}, ErrNotFound
		}
		return RepoMapping{}, err
	}
	m.CreatedAt = parseTime(created)
	m.UpdatedAt = parseTime(updated)
	return m, nil
}

func parseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return t
}
