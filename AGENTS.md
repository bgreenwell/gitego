# AGENTS.md for git-ego

@../CLAUDE.md

`git-ego` is a command-line tool that solves the common problem of managing
multiple git identities. It allows you to define separate profiles for work,
personal projects, and clients, and then automatically switch between them
based on your working directory.

The tool manages `user.name`, `user.email`, SSH keys, and personal access
tokens (PATs), acting as a unified and intelligent manager for your git
identity. It is built on native git features like `includeif` and credential
helpers, ensuring it works seamlessly without fighting against git's own
mechanisms.

## Architecture

The project is organized into three main packages, following standard Go
practices:

- **`cmd/`**: All CLI logic. Each command is its own file (e.g., `add.go`,
  `list.go`). Uses `spf13/cobra`; `root.go` sets up the main command and the
  save-then-reconcile workflow. `resolution.go` is the shared safety model for
  `status`, the credential helper, and the commit hook: Git's effective
  `gitego.profile` is authoritative, while repository-root `.gitego` files and
  the most-specific auto-rule are expectations.
- **`config/`**: Loading, saving, and managing user profiles and application
  settings, including secure keychain interaction. `config.go` defines the
  data structures for profiles and auto-activation rules, serialized with
  `gopkg.in/yaml.v3`. YAML is authoritative. `safety.go` validates desired
  state, generates credential IDs, and serializes mutations with
  `~/.gitego/config.lock`. `atomic.go` and the platform replacement/sync files
  publish complete files safely. `reconcile.go` regenerates profile, default,
  and conditional include files and maintains one final managed include in
  `~/.gitconfig`. `keyring.go` manages credential-ID-keyed secure storage via
  `github.com/zalando/go-keyring`.
- **`utils/`**: Helper functions that interact with the git CLI and file
  system. `git.go` reads effective values and their origins, discovers the
  repository root, and applies repository-local Git configuration.

Mutating commands must acquire the configuration lock, validate and atomically
save YAML first, and then reconcile derived Git files. If reconciliation fails,
leave the valid YAML in place and return actionable `doctor --repair` guidance.
Do not use `.gitego` to select identity or credentials; it is an assertion only.

## Main dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/gofrs/flock` — cross-process mutation locking
- `github.com/zalando/go-keyring` — cross-platform credential store access
- `gopkg.in/yaml.v3` — YAML parsing/emitting

## Verification and releases

- Before finishing Go changes, run the checks in `../CLAUDE.md`: `gofmt -s`,
  `go vet ./...`, `golangci-lint run`, and `go test -v ./...`.
- For safety, persistence, credential, or release changes, also run shuffled
  repeated tests, race tests, `govulncheck`, Linux/macOS/Windows builds for
  amd64 and arm64, and the Nix build. Use the golangci-lint version pinned in
  `.github/workflows/go-ci.yml`.
- Release builds and CI are defined in `.github/workflows/`. Package publishing
  updates the Homebrew tap and Scoop bucket automatically. WinGet requires a
  one-time manually reviewed manifest PR before the `winget.yml` update
  workflow can publish later versions.
- Package workflows rely on repository secrets named `HOMEBREW_TAP_TOKEN`,
  `SCOOP_BUCKET_TOKEN`, and `WINGET_TOKEN`. Never place token values in source,
  workflow logs, or documentation.
