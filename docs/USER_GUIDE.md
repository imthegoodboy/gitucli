# gituCli User Guide

This guide explains how to use `gituCli` from zero: install it, create SSH keys, add keys to GitHub, connect each project to the correct GitHub account, and check that everything is safe.

## What gituCli Does

`gituCli` helps when you use multiple GitHub accounts on one laptop.

Example:

- Project A should commit and push as GitHub account A.
- Project B should commit and push as GitHub account B.
- They must not mix emails, SSH keys, remotes, or contribution history.

gituCli handles this by setting each project with:

- one local Git name
- one local Git email
- one SSH key
- one SSH alias
- one GitHub remote
- safety hooks before commit and push

## Before You Start

You need:

- Git installed
- OpenSSH / `ssh-keygen` installed
- Go installed only if you want to build from source
- a GitHub account for each identity you want to use
- access to the GitHub repository you want to push to

On this Windows machine, Git and `ssh-keygen` are already available. Go is installed at:

```powershell
C:\Program Files\Go\bin\go.exe
```

## 1. Build gituCli

From the project folder:

```powershell
cd C:\Users\parth\Desktop\gituCli
& 'C:\Program Files\Go\bin\go.exe' build -o .\bin\gitu.exe .\cmd\gitu
```

Now test the binary:

```powershell
.\bin\gitu.exe
```

You should see the command menu.

## 2. Understand SSH Keys

An SSH key has two files:

```text
private key: C:\Users\you\.ssh\gitu_startup
public key:  C:\Users\you\.ssh\gitu_startup.pub
```

Important:

- The private key stays on your laptop.
- Never paste the private key into GitHub.
- The public key, ending in `.pub`, is the one you add to GitHub.

GitHub’s official SSH docs are here:

- [Connecting to GitHub with SSH](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)
- [Adding a new SSH key to your GitHub account](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account)

## 3. Add Your First Profile

A profile means one GitHub identity.

Example for a startup account:

```powershell
.\bin\gitu.exe profile add `
  --name startup `
  --github-user startupAccount `
  --git-name "Startup Account" `
  --email founder@startup.com `
  --key "$env:USERPROFILE\.ssh\gitu_startup" `
  --alias github-startup
```

What each field means:

- `--name startup`: local nickname inside gituCli
- `--github-user startupAccount`: GitHub username
- `--git-name "Startup Account"`: name used in commits
- `--email founder@startup.com`: email used in commits
- `--key ...\.ssh\gitu_startup`: private SSH key path
- `--alias github-startup`: SSH alias used by repo remotes

The email must be verified on that GitHub account. GitHub contributions depend on commit email.

## 4. Generate The SSH Key

After creating the profile:

```powershell
.\bin\gitu.exe key generate startup
```

This creates:

```text
C:\Users\you\.ssh\gitu_startup
C:\Users\you\.ssh\gitu_startup.pub
```

If you want to see the public key:

```powershell
Get-Content "$env:USERPROFILE\.ssh\gitu_startup.pub"
```

Copy the full output. It usually starts with:

```text
ssh-ed25519
```

## 5. Add The Public Key To GitHub

Do this for the matching GitHub account only.

1. Open GitHub in your browser.
2. Log in as the account for this profile, for example `startupAccount`.
3. Click your profile photo.
4. Open `Settings`.
5. Go to `SSH and GPG keys`.
6. Click `New SSH key` or `Add SSH key`.
7. Give it a clear title, for example `gituCli startup laptop`.
8. Choose authentication key if GitHub asks for the key type.
9. Paste the `.pub` key content.
10. Click `Add SSH key`.

Do not add the same key to the wrong GitHub account. One profile should use one GitHub account.

## 6. Initialize A Project

Suppose your repo is:

```text
C:\projects\ai-saas
```

And the GitHub repo is:

```text
startupAccount/ai-saas
```

Run:

```powershell
.\bin\gitu.exe init C:\projects\ai-saas `
  --profile startup `
  --repo startupAccount/ai-saas
```

gituCli will:

- create or use the Git repo
- set local `user.name`
- set local `user.email`
- write/update the safe SSH config block
- set the remote to `git@github-startup:startupAccount/ai-saas.git`
- install safety hooks
- store the repo mapping in SQLite

## 7. Validate The Project

Run:

```powershell
.\bin\gitu.exe validate C:\projects\ai-saas
```

Expected output:

```text
[OK] C:\projects\ai-saas is identity-safe for profile startup
```

You can also manually check:

```powershell
git -C C:\projects\ai-saas config --local --get user.email
git -C C:\projects\ai-saas remote get-url origin
```

Expected:

```text
founder@startup.com
git@github-startup:startupAccount/ai-saas.git
```

## 8. Push Normally

After setup, use normal Git:

```powershell
cd C:\projects\ai-saas
git add .
git commit -m "update project"
git push
```

The safety hooks check the identity before commit and push.

## 9. Add A Second GitHub Account

Example for a personal account:

```powershell
.\bin\gitu.exe profile add `
  --name personal `
  --github-user personalAccount `
  --git-name "Personal Account" `
  --email personal@example.com `
  --key "$env:USERPROFILE\.ssh\gitu_personal" `
  --alias github-personal

.\bin\gitu.exe key generate personal
Get-Content "$env:USERPROFILE\.ssh\gitu_personal.pub"
```

Add `gitu_personal.pub` to the `personalAccount` GitHub account.

Then initialize another repo:

```powershell
.\bin\gitu.exe init C:\projects\portfolio `
  --profile personal `
  --repo personalAccount/portfolio
```

Now:

- `C:\projects\ai-saas` uses startup identity
- `C:\projects\portfolio` uses personal identity

They can be committed and pushed in parallel without mixing accounts.

## 10. Fix A Broken Project

If something is wrong:

```powershell
.\bin\gitu.exe validate C:\projects\ai-saas
```

If it reports repairable issues:

```powershell
.\bin\gitu.exe repair C:\projects\ai-saas
```

Repair restores:

- local Git name
- local Git email
- remote URL
- hooks
- managed SSH config block

## 11. Use The Daemon

Run one check:

```powershell
.\bin\gitu.exe daemon --once
```

Run continuous checks:

```powershell
.\bin\gitu.exe daemon
```

The daemon only checks repos already initialized by gituCli. It does not scan your whole laptop.

## 12. Safe Dummy Test

To test without touching your real SSH config, use:

```powershell
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ('gitu-test-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $temp | Out-Null
$env:GITU_HOME = Join-Path $temp 'home'
$env:GITU_SSH_CONFIG = Join-Path $temp 'ssh\config'
$db = Join-Path $temp 'gitu.db'

$repo = Join-Path $temp 'repo'
$key = Join-Path $temp 'test-key'

.\bin\gitu.exe --db $db profile add `
  --name test `
  --github-user testUser `
  --git-name "Test User" `
  --email test@example.com `
  --key $key `
  --alias github-test

.\bin\gitu.exe --db $db key generate test
.\bin\gitu.exe --db $db init $repo --profile test --repo owner/test-repo
.\bin\gitu.exe --db $db validate $repo
```

This creates a temporary repo, temporary database, temporary SSH key, and temporary SSH config.

## Common Problems

### Push says permission denied

Check:

- Did you add the correct `.pub` key to the correct GitHub account?
- Does that GitHub account have access to the repo?
- Does `gitu validate` pass?

### Commit is blocked

Run:

```powershell
.\bin\gitu.exe validate C:\path\to\repo
.\bin\gitu.exe repair C:\path\to\repo
```

### Contributions show on the wrong GitHub account

Check the commit email:

```powershell
git -C C:\path\to\repo config --local --get user.email
```

That email must be verified on the intended GitHub account.

### I moved gitu.exe

Run repair for each managed repo:

```powershell
.\bin\gitu.exe repair C:\path\to\repo
```

This rewrites the hook scripts to point at the new binary path.

## Daily Workflow

After setup, normal daily use is simple:

```powershell
cd C:\path\to\repo
git status
git add .
git commit -m "message"
git push
```

Use gituCli when adding a new repo, adding a new identity, validating, or repairing.

