# gituCli — Multi GitHub Identity Manager CLI

## What is gituCli?

`gituCli` is a smart CLI tool that lets developers manage multiple GitHub accounts and repositories on a single laptop without conflicts.

It automatically:

* isolates GitHub identities per project
* manages SSH keys
* manages Git commit emails
* prevents contribution conflicts
* auto-configures repository remotes
* validates identity before push/commit
* works silently in the background

The goal is:

```text id="42wshg"
One project = One GitHub identity
```

without:

* repeated login/logout
* token confusion
* wrong commits
* wrong contributions
* SSH conflicts

---

# The Main Problem gituCli Solves

Most developers face this issue:

```text id="5b7xpv"
Project A accidentally pushes using GitHub Account B
```

Result:

* wrong contribution graph
* wrong profile avatar
* organization access issues
* private repo permission failures
* mixed commit history

gituCli solves this by creating:

```text id="bl2xud"
Per-project isolated Git environments
```

---

# Real World Example (5 Projects)

---

# Example Setup

| Project              | GitHub Account | Purpose      |
| -------------------- | -------------- | ------------ |
| AI SaaS              | Account A      | Startup      |
| Open Source Tool     | Account B      | Public OSS   |
| Client Dashboard     | Account C      | Freelance    |
| Personal Portfolio   | Account D      | Personal     |
| Secret Research Repo | Account E      | Experimental |

Normally this creates chaos.

With gituCli:

```text id="w4v5vk"
Each project gets:
- unique SSH identity
- unique git email
- unique git remote
```

fully isolated.

---

# Example Workflow

# Project 1 — AI SaaS

User runs:

```bash id="9fapb7"
gitu init
```

CLI asks:

```text id="t9ehc5"
Select repo directory:
```

User selects:

```text id="m9lq9f"
~/projects/ai-saas
```

CLI asks:

```text id="35wkl2"
GitHub username:
startupAccount
```

```text id="vg9qzh"
GitHub email:
founder@startup.com
```

```text id="ezl8po"
SSH key path:
~/.ssh/startup_key
```

Then gituCli configures:

```text id="yv08mx"
git config user.email founder@startup.com
```

and creates SSH alias:

```text id="yik8o8"
github-startup
```

Now every push inside this repo:

* uses startup GitHub
* uses startup contributions
* uses startup SSH key

---

# Project 2 — Open Source

Different project.

Different identity.

Different contributions.

Same laptop.

No conflict.

---

# Project 3 — Client Project

Uses:

* client email
* client SSH key
* client GitHub org

No interference with startup repos.

---

# Project 4 — Personal Portfolio

Uses:

* personal GitHub profile
* personal commits

Still isolated.

---

# Project 5 — Research Repo

Can even use:

* anonymous identity
* different signing key
* different GitHub enterprise

All simultaneously.

---

# How gituCli Actually Works Internally

# CORE IDEA

gituCli combines:

```text id="bhwp89"
Git local config
+
SSH alias routing
+
Repository mapping
```

---

# System Architecture

```text id="9esw88"
┌────────────────────┐
│     gituCli CLI    │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ Repository Scanner │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ Identity Manager   │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ Git Config Engine  │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ SSH Config Router  │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ GitHub Authentication │
└────────────────────┘
```

---

# Background Working

# 1. Local Git Config Isolation

Inside every repo:

```bash id="owbx5r"
.git/config
```

gituCli writes:

```ini id="lqocv0"
[user]
    name = StartupAccount
    email = founder@startup.com
```

This ensures:

* commits belong to correct GitHub account
* contribution graph stays correct

---

# 2. SSH Identity Isolation

gituCli manages:

```bash id="1sm4n4"
~/.ssh/config
```

Example:

```config id="hv30vg"
Host github-startup
    HostName github.com
    User git
    IdentityFile ~/.ssh/startup_key
```

This routes:

* startup repo → startup SSH key

---

# 3. Remote URL Rewriting

Original remote:

```bash id="nlg7yq"
git@github.com:user/repo.git
```

gituCli rewrites:

```bash id="jny6z5"
git@github-startup:user/repo.git
```

Now SSH automatically selects correct identity.

---

# 4. Contribution Isolation

This is VERY important.

GitHub contributions depend on:

```text id="y7y3f5"
commit email
```

NOT SSH key.

So gituCli always ensures:

```text id="v1txh8"
Correct repo → correct email
```

This prevents:

* wrong contribution graphs
* mixed identities
* commit mismatches

---

# User Perspective (UX)

# User Experience

User opens terminal:

```bash id="6vwpd9"
gitu
```

Beautiful colorful ASCII interface appears:

```text id="rt3h2k"
██████╗ ██╗████████╗██╗   ██╗
██╔══██╗██║╚══██╔══╝██║   ██║
██║  ██║██║   ██║   ██║   ██║
██║  ██║██║   ██║   ██║   ██║
██████╔╝██║   ██║   ╚██████╔╝
╚═════╝ ╚═╝   ╚═╝    ╚═════╝

Multi GitHub Identity Manager
```

Then menu:

```text id="b5d93t"
[1] Initialize Project
[2] Switch Identity
[3] Validate Repo
[4] List Profiles
[5] Generate SSH Key
[6] Fix Repo Issues
[7] Contribution Safety Check
```

---

# Example User Flow

```bash id="eg2htk"
gitu init
```

CLI auto:

* scans repo
* detects git repo
* asks GitHub identity
* configures SSH
* rewrites remote
* validates email
* creates hooks

Done.

Now user simply:

```bash id="8g1lqf"
git push
```

No extra steps.

---

# CLI Features

# Core Features

| Feature                 | Purpose                     |
| ----------------------- | --------------------------- |
| Identity Isolation      | Separate GitHub accounts    |
| SSH Manager             | Auto SSH setup              |
| Remote Rewriter         | Correct repo routing        |
| Contribution Protection | Prevent wrong commits       |
| Repo Validator          | Detect configuration issues |
| Git Hook Protection     | Block invalid commits       |
| Multi Profile Storage   | Save multiple identities    |
| Auto Repair             | Fix broken repo config      |
| Interactive UI          | Friendly CLI experience     |
| Background Validation   | Silent safety checks        |

---

# Advanced Features

# 1. Smart Auto Detection

gituCli can detect:

* repo owner
* org repo
* mismatched emails
* invalid SSH keys

---

# 2. Contribution Safety Engine

Before commit:

```text id="6rx17d"
Does this repo use correct identity?
```

If not:

BLOCK COMMIT.

---

# 3. Background Repo Watcher

Optional daemon:

```bash id="ndmu9h"
gitu daemon
```

Monitors:

* repo changes
* config corruption
* SSH issues

---

# 4. Team Profiles

```bash id="b6m7wq"
gitu profile add startup
```

Switch identities instantly.

---

# 5. Secure Storage

Store:

* profile metadata
* repo mappings

inside:

* SQLite
* encrypted config

Never store passwords.

---

# CLI Tech Stack

# Language

Go

Why?

* single binary
* fast
* cross-platform
* easy CLI tooling
* good process execution

---

# Recommended Libraries

| Purpose        | Library                |
| -------------- | ---------------------- |
| CLI Framework  | `cobra`                |
| Interactive UI | `bubbletea`            |
| ASCII Styling  | `lipgloss`             |
| Config         | `viper`                |
| SQLite         | `modernc.org/sqlite`   |
| Git Operations | `go-git` or shell exec |
| SSH            | native ssh tooling     |

---

# Suggested Folder Structure

```text id="i1mrf8"
gituCli/
│
├── cmd/
├── internal/
│   ├── git/
│   ├── ssh/
│   ├── profiles/
│   ├── hooks/
│   ├── validator/
│   ├── storage/
│   ├── remote/
│   └── ui/
│
├── configs/
├── assets/
└── scripts/
```

---

# Background System Working

# When User Runs git push

Actual hidden flow:

```text id="f1u2y0"
git push
   ↓
Repo remote checked
   ↓
SSH alias detected
   ↓
Correct SSH key selected
   ↓
GitHub authenticates
   ↓
Commit email verified
   ↓
Push succeeds
```

Everything automatic.

---

# Git Hook System

Very important feature.

Inside:

```text id="ck63pk"
.git/hooks/
```

gituCli installs:

```text id="c75pg1"
pre-commit
pre-push
```

Hooks verify:

* correct email
* correct SSH alias
* correct remote

If mismatch:

```text id="xy7x7s"
ERROR:
Wrong GitHub identity detected.
```

This prevents accidental contribution conflicts forever.

---

# ASCII + Colorful CLI Design

Use:

* gradient colors
* animated loaders
* ASCII banners
* progress spinners
* repo cards
* status icons

Example:

```text id="by7c6j"
✓ SSH Configured
✓ Git Email Valid
✓ Remote Verified
✓ Contribution Safe
```

Using:

* `bubbletea`
* `lipgloss`

you can create a modern terminal UI similar to:

* lazygit
* gh CLI
* Warp terminal tools

---

# VERY IMPORTANT THINGS TO BE CAREFUL ABOUT

# 1. NEVER Use Global Git Config

BAD:

```bash id="5e2dwt"
git config --global
```

This breaks identity isolation.

Always use:

* local repo config only

---

# 2. Never Store Passwords

Only use:

* SSH keys
* optionally GitHub PAT

Never store:

* GitHub passwords

---

# 3. Contribution Mapping Depends on Email

Most important rule.

Wrong email = wrong contributions.

Always validate:

* repo email
* GitHub verified email

---

# 4. SSH Alias Must Be Unique

Avoid:

```text id="sln8jf"
Host github
```

Use:

```text id="3wm5kp"
Host github-startup
Host github-client
```

---

# 5. Remote URL Must Match Alias

Wrong:

```bash id="r5suh6"
git@github.com:
```

Correct:

```bash id="g07l41"
git@github-startup:
```

---

# 6. Prevent Hook Corruption

Users may delete:

* `.git/hooks`

So gituCli should:

* revalidate hooks
* auto-reinstall

---

# 7. Handle Existing Repos Carefully

Many repos already have:

* remotes
* SSH configs
* user.email

Your tool must:

* backup old configs
* rollback safely

---

# 8. SSH Config Parsing Is Sensitive

Avoid overwriting:

* entire SSH config

Only append/update controlled blocks.

---

# 9. Never Modify System SSH Settings Aggressively

Only manage:

* gituCli-generated aliases

Do not destroy existing SSH configs.

---

# 10. Always Provide Repair Commands

Example:

```bash id="mxjlwm"
gitu repair
```

to fix:

* broken remotes
* invalid SSH keys
* email mismatch

---

# Final Vision of gituCli

gituCli becomes:

```text id="rn7r2u"
The identity layer above Git.
```

A smart system that:

* isolates GitHub accounts
* protects contributions
* automates SSH
* prevents mistakes
* simplifies multi-account development

all from one beautiful CLI.
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             