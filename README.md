# gituCli

`gituCli` is a Go CLI for using multiple GitHub accounts on one laptop without mixing commits, SSH keys, or remotes.

The rule is simple:

```text
one project = one GitHub identity
```

For every initialized repo, `gitu` manages:

- repo-local `user.name` and `user.email`
- a unique SSH host alias in the managed block of `~/.ssh/config`
- the GitHub remote URL, rewritten to that alias
- strict `pre-commit` and `pre-push` hooks
- a SQLite repo-to-profile mapping

No passwords or GitHub tokens are stored.

## Install

Build the binary:

```powershell
go build -o .\bin\gitu.exe .\cmd\gitu
```

Add `bin` to your PATH, or call the binary by absolute path.

## Quick Start

Run the guided app:

```powershell
gitu
```

The menu walks you through setup, profiles, repo validation, repair, autocommit, and daemon checks one question at a time.

You can still use direct commands when scripting or moving fast.

Create or attach a profile and configure a repo:

```powershell
gitu init C:\path\to\project --profile startup --repo owner/repo --generate-key
```

If the profile does not exist, `gitu` asks for:

- GitHub username
- Git commit name
- Git commit email
- SSH private key path
- SSH alias

After setup, normal Git commands stay normal:

```powershell
git commit -m "change"
git push
```

The repo will push through the configured SSH alias, and commits will use the configured local email.

## Common Commands

```powershell
gitu profile add
gitu profile list
gitu key generate startup
gitu validate C:\path\to\project
gitu repair C:\path\to\project
gitu autocommit C:\path\to\project --message "checkpoint"
gitu autocommit C:\path\to\project --message "checkpoint" --interval 30m
gitu daemon
```

## Two Accounts On One Laptop

```powershell
gitu profile add --name startup --github-user startupUser --email founder@startup.com
gitu init C:\projects\ai-saas --profile startup --repo startupUser/ai-saas --generate-key

gitu profile add --name personal --github-user personalUser --email me@users.noreply.github.com
gitu init C:\projects\portfolio --profile personal --repo personalUser/portfolio --generate-key
```

Each repo gets its own local Git email and its own SSH alias remote. You can commit and push in both repos without logging in and out.

## Safety Notes

- `gitu` never writes global Git config.
- GitHub contribution attribution depends on commit email, not SSH key.
- You must add the generated `.pub` key to the matching GitHub account.
- The daemon watches only repos already initialized by `gitu`.

## More Docs

- `docs/USER_GUIDE.md`: step-by-step beginner guide for setup, SSH keys, GitHub, repos, validate, repair, and daily use.
- `docs/AUTOCOMMIT.md`: timed and repeated safe auto-commit workflow.
- `docs/IDENTITY_MODEL.md`: how contribution identity and SSH routing work.
- `docs/OPERATIONS.md`: validate, repair, daemon, and troubleshooting.
- `docs/END_TO_END_FLOW.md`: safe dummy-project test flow.
- `docs/AUDIT_REPORT.md`: latest audit checklist and verification summary.
