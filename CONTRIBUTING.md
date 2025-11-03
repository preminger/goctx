# Contributing

The project is not yet open to external contributions; but stay tuned!

## Development workflow notes

- Run `make init` once after cloning. This installs dev tools and configures Git to use the repo-local hooks in `.githooks/`.
- A pre-push hook enforces that any branch push (e.g., for a PR) includes a change to `CHANGELOG.md`. Pushes to `main`/`master` and tag pushes are exempt.
- If you must bypass the check (e.g., hotfix or infrastructure changes), export `BYPASS_CHANGELOG_CHECK=1` for that push:
  
  ```bash
  BYPASS_CHANGELOG_CHECK=1 git push
  ```
