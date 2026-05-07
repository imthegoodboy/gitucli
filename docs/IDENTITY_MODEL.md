# gituCli Identity Model

GitHub identity has two separate parts:

- **Commit attribution:** controlled by the repo-local Git `user.email`.
- **Push authentication:** controlled by the SSH key selected by the repo remote host alias.

Both must match the intended GitHub account.

## Profile

A gitu profile stores metadata only:

- profile name
- GitHub username
- Git commit name
- Git commit email
- SSH private key path
- SSH alias

No passwords or tokens are stored.

## Repo Mapping

Each initialized repo has one mapping in SQLite:

- absolute repo path
- selected profile
- managed remote name, usually `origin`
- original remote URL backup
- expected alias remote
- strict hook mode

This mapping is the source of truth for validation and repair.

## Git Config

gitu writes only local repo config:

```powershell
git config --local user.name  "Name"
git config --local user.email "email@example.com"
```

Global Git config is intentionally ignored.

## SSH Alias Routing

gitu writes a managed block in `~/.ssh/config`:

```sshconfig
# >>> gituCli managed identities >>>
Host github-startup
    HostName github.com
    User git
    IdentityFile C:/Users/me/.ssh/gitu_startup
    IdentitiesOnly yes
# <<< gituCli managed identities <<<
```

User-owned SSH config outside the markers is preserved.

## Remote Rewriting

GitHub remotes are rewritten to the selected alias:

```text
git@github.com:owner/repo.git
```

becomes:

```text
git@github-startup:owner/repo.git
```

That causes SSH to choose the profile-specific key.

