# Auto Commit

`gitu autocommit` safely commits changes for a managed repo after checking that the repo identity is correct.

It is useful when you want automatic checkpoints while working, but still want protection against the wrong GitHub account.

## Safety Rules

Before every auto commit, gituCli runs the normal identity validation:

- repo must be initialized by gituCli
- local Git name must match the mapped profile
- local Git email must match the mapped profile
- remote must use the mapped SSH alias
- SSH key must exist
- managed hooks must exist

If validation fails, autocommit stops and does not commit.

## Commit Once Now

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "daily checkpoint"
```

This stages all changes with:

```powershell
git add -A
```

Then commits with your message.

## Dry Run

Check what would be committed:

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "checkpoint" --dry-run
```

Dry run validates identity and prints pending changes, but does not stage or commit.

## Commit At A Specific Time

Use local 24-hour `HH:MM` time:

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "night checkpoint" --at 22:30
```

If the time already passed today, gituCli waits until that time tomorrow.

While waiting, the CLI shows a small animated wait indicator so you can tell the scheduler is still alive.

## Commit Repeatedly

Run every 30 minutes:

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "work checkpoint" --interval 30m
```

Run every 2 hours:

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "auto save" --interval 2h
```

This is a foreground process. Keep that terminal open while you want the schedule to run.

## Wait Until A Time, Then Repeat

Start at 09:30, then commit every hour:

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "hourly checkpoint" --at 09:30 --interval 1h
```

## Commit And Push

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "checkpoint" --push
```

Only use `--push` after you have added the profile public key to the matching GitHub account and confirmed normal `git push` works.

## Default Message

If you do not pass `--message`, gituCli uses a timestamped message:

```text
auto commit 2026-05-07 08:30:00
```

## Good Usage Pattern

For active coding sessions:

```powershell
.\bin\gitu.exe validate C:\path\to\repo
.\bin\gitu.exe autocommit C:\path\to\repo --message "checkpoint" --interval 30m
```

For one final checkpoint:

```powershell
.\bin\gitu.exe autocommit C:\path\to\repo --message "final checkpoint"
```

## Important Notes

- Autocommit commits all tracked and untracked changes in the repo.
- It does not bypass identity validation.
- It does not store secrets.
- It does not run in the background after the terminal closes.
- If there are no changes, it prints a skip message and does nothing.
