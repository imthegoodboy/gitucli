# gituCli Audit Report

Date: 2026-05-07

## Summary

The current build was audited against the core promise:

```text
one repo = one GitHub identity
```

The audit found and fixed one real bug: the persistent `--db` flag was not reaching subcommands because command environment values were copied too early. A regression test now proves `--db` works for subcommands.

## Production Checks

Verified:

- Go project builds successfully.
- Unit tests pass across CLI, core service, remote parsing, SSH config, and SQLite storage.
- Autocommit dry-run and commit flow pass in a dummy repo.
- Project-local agent skill validates.
- CLI root menu exits cleanly in non-interactive terminals.
- CLI output has clearer colored statuses for success, failure, commands, prompts, and guard errors.
- Profile input now validates malformed emails and unsafe SSH aliases.
- Managed SSH config writes are isolated to the marked gituCli block.
- Repo-local Git config is used; global Git config is not written.

## End-to-End Dummy Flow

The dummy flow used:

- isolated SQLite database via `--db`
- isolated app home via `GITU_HOME`
- isolated SSH config via `GITU_SSH_CONFIG`
- two temporary Git repos
- two profiles
- two generated SSH keys

Verified:

- repo A uses `account-a@example.com`
- repo A remote is `git@github-account-a:owner/repo-a.git`
- repo B uses `account-b@example.com`
- repo B remote is `git@github-account-b:owner/repo-b.git`
- guard blocks repo A after intentionally changing its local email to `wrong@example.com`
- repair restores repo A to the correct email
- daemon `--once` reports both configured repos as OK

## Advanced Feature Audit

Added and verified `gitu autocommit`:

- validates repo identity before committing
- supports `--message`
- supports `--dry-run`
- supports `--at HH:MM`
- supports repeated `--interval`
- supports optional `--push`
- skips clean repos without creating empty commits
- preserves custom `--db` paths inside installed Git hook scripts

## Commands Run

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./...
& 'C:\Program Files\Go\bin\go.exe' build -o .\bin\gitu.exe .\cmd\gitu
python C:\Users\parth\.codex\skills\.system\skill-creator\scripts\quick_validate.py .\.agents\skills\gitucli-identity-guard
```

The complete reproducible dummy-project procedure is in `docs/END_TO_END_FLOW.md`.

## Remaining Operational Notes

- Real GitHub pushes still require adding each generated `.pub` key to the matching GitHub account.
- GitHub contribution correctness depends on using an email verified on that account.
- Moving `gitu.exe` after repo initialization requires `gitu repair` so hooks point to the new binary location.
