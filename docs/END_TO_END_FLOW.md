# End-to-End Dummy Project Flow

This flow verifies gituCli without touching your real SSH config or normal gitu database.

## 1. Build

```powershell
go test ./...
go build -o .\bin\gitu.exe .\cmd\gitu
```

If `go` is not on PATH but Go is installed in `C:\Program Files\Go`, use:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./...
& 'C:\Program Files\Go\bin\go.exe' build -o .\bin\gitu.exe .\cmd\gitu
```

## 2. Create an Isolated Test Area

```powershell
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ('gitu-e2e-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $temp | Out-Null
$env:GITU_HOME = Join-Path $temp 'home'
$env:GITU_SSH_CONFIG = Join-Path $temp 'ssh\config'
$db = Join-Path $temp 'gitu.db'
```

`GITU_HOME` isolates the default database, and `GITU_SSH_CONFIG` isolates the SSH config file.

## 3. Configure Account A

```powershell
$repoA = Join-Path $temp 'repo-a'
$keyA = Join-Path $temp 'account-a-key'

.\bin\gitu.exe --db $db profile add `
  --name account-a `
  --github-user accountA `
  --git-name 'Account A' `
  --email account-a@example.com `
  --key $keyA `
  --alias github-account-a

.\bin\gitu.exe --db $db key generate account-a
.\bin\gitu.exe --db $db init $repoA --profile account-a --repo owner/repo-a
.\bin\gitu.exe --db $db validate $repoA
```

Expected:

- validation returns `[OK]`
- `git -C $repoA config --local --get user.email` returns `account-a@example.com`
- `git -C $repoA remote get-url origin` returns `git@github-account-a:owner/repo-a.git`

## 4. Configure Account B

```powershell
$repoB = Join-Path $temp 'repo-b'
$keyB = Join-Path $temp 'account-b-key'

.\bin\gitu.exe --db $db profile add `
  --name account-b `
  --github-user accountB `
  --git-name 'Account B' `
  --email account-b@example.com `
  --key $keyB `
  --alias github-account-b

.\bin\gitu.exe --db $db key generate account-b
.\bin\gitu.exe --db $db init $repoB --profile account-b --repo owner/repo-b
.\bin\gitu.exe --db $db validate $repoB
```

Expected:

- repo A still uses `account-a@example.com` and `github-account-a`
- repo B uses `account-b@example.com` and `github-account-b`
- the two repos do not share local Git identity or remote aliases

## 5. Prove The Guard Blocks Drift

```powershell
git -C $repoA config --local user.email wrong@example.com
.\bin\gitu.exe --db $db guard pre-commit --repo $repoA
```

Expected:

- guard exits non-zero
- output says the repo identity is unsafe

Repair it:

```powershell
.\bin\gitu.exe --db $db repair $repoA
.\bin\gitu.exe --db $db validate $repoA
```

Expected:

- validation returns `[OK]`
- repo A email is restored to `account-a@example.com`

## 6. Daemon Sweep

```powershell
.\bin\gitu.exe --db $db daemon --once
```

Expected:

- configured repos print `[OK]`
- missing managed hooks are restored automatically

## What This Test Does Not Prove

This dummy flow does not push to GitHub. For real pushes, add each generated `.pub` key to its matching GitHub account first.
