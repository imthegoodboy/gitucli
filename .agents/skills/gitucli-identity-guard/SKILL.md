---
name: gitucli-identity-guard
description: Guide for modifying gituCli, a Go CLI that isolates multiple GitHub accounts per repository. Use when working on gituCli identity profiles, local Git config, SSH alias routing, remote URL rewriting, Git hooks, validation, repair, daemon behavior, tests, or documentation.
---

# gituCli Identity Guard

## Overview

Use this skill to preserve the central safety promise of gituCli: one repository maps to exactly one GitHub identity. The identity is enforced through repo-local Git author config, a managed SSH alias, a rewritten remote URL, strict Git hooks, and SQLite metadata.

## Core Rules

- Never add `git config --global` behavior. gituCli must only write repo-local Git config.
- Never store GitHub passwords or personal access tokens. v1 uses SSH keys and metadata only.
- Treat GitHub contribution attribution as email-driven. The configured `user.email` must match the selected profile.
- Preserve user-owned SSH config. Only replace the marked `gituCli` block in `~/.ssh/config`.
- Keep SSH aliases unique and profile-specific, such as `github-work` or `github-client`.
- Rewrite managed GitHub remotes to `git@<alias>:owner/repo.git`.
- Hooks must block by default on identity mismatch.

## Implementation Workflow

1. Identify whether the change affects profiles, repo mappings, Git config, SSH config, remotes, hooks, validation, repair, or docs.
2. Keep behavior centered in the service layer before adding CLI surface area.
3. Add focused tests for every safety-sensitive branch, especially multi-repo and multi-profile cases.
4. Verify no global Git config writes were introduced.
5. Update docs when user-visible setup, safety guarantees, or troubleshooting changes.

## Testing Expectations

- Test remote parsing and rewrite behavior for SSH, HTTPS, and existing alias remotes.
- Test SSH config rendering so user blocks remain untouched and managed blocks are replaced atomically.
- Test storage with temporary SQLite databases.
- Test integration flows with temporary Git repositories whenever behavior crosses Git config, remotes, hooks, and storage.
- For guard changes, include failure cases that prove commits or pushes are blocked when the profile email or remote alias drifts.

## Documentation Pointers

- Read `README.md` for normal user workflows.
- Read `docs/IDENTITY_MODEL.md` before changing contribution, SSH, or remote behavior.
- Read `docs/OPERATIONS.md` before changing repair, daemon, or troubleshooting behavior.
