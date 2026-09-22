# Contributing to Fortify

Thanks for your interest in contributing! 🏰

## Golden rule

**Never commit directly to `main`.** `main` is the released branch and must stay stable and deployable at all times. **Every** change — feature, fix, or docs — goes through its own branch and a pull request. `main` is branch-protected, so a pull request is required to merge.

## Workflow

1. **Fork** the repo (external contributors) — collaborators can branch directly.
2. **Branch off `main`**, one branch per issue/feature, named by intent (`feat/…`, `fix/…`, `docs/…`):
   ```bash
   git checkout main && git pull
   git checkout -b feat/your-feature
   ```
3. **Commit** small, focused changes with clear messages.
4. **Push** and open a **pull request** into `main`. Reference the issue it closes (e.g. "Closes #9").
5. Keep `main` releasable — don't merge half-finished or failing work.

## Code quality

Please keep code clean and tested before submitting.

## Responsible use

Fortify is a security tool. Only test systems you **own** or have **explicit written permission** to test — see [Responsible Use](README.md#responsible-use) in the README.
