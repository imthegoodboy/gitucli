# gituCli Operations

## Validate

Run:

```powershell
gitu validate C:\path\to\repo
```

Validation checks:

- repo mapping exists
- mapped profile exists
- repo-local `user.name` and `user.email`
- managed remote URL and SSH alias
- SSH key path exists
- `pre-commit` and `pre-push` hooks are managed by gitu

## Repair

Run:

```powershell
gitu repair C:\path\to\repo
```

Repair restores:

- local Git author config
- managed remote URL
- managed Git hooks
- managed SSH config block

It does not create missing GitHub-side SSH key registrations. Add the `.pub` key to the correct GitHub account yourself.

## Daemon

Run:

```powershell
gitu daemon
```

The daemon watches only repos initialized by `gitu`. It validates each configured repo and restores missing managed hooks. It does not scan the full laptop.

For one foreground sweep:

```powershell
gitu daemon --once
```

## Windows Notes

- Git hooks are POSIX shell scripts executed by Git for Windows.
- Hook scripts call the installed `gitu.exe` path captured during `gitu init` or `gitu repair`.
- If you move the binary, run `gitu repair` for each repo so hooks point to the new binary path.

## Troubleshooting

If a commit or push is blocked:

1. Run `gitu validate`.
2. Run `gitu repair` if the issue is repairable.
3. Confirm the profile email is verified on the intended GitHub account.
4. Confirm the profile `.pub` key is added to that GitHub account.

