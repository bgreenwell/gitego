<p align="center">
  <img src="assets/logo.svg" alt="git-ego" width="620" />
</p>

[![CI](https://img.shields.io/github/actions/workflow/status/bgreenwell/git-ego/go-ci.yml?style=for-the-badge)](https://github.com/bgreenwell/git-ego/actions/workflows/go-ci.yml)
[![Go Reference](https://img.shields.io/badge/go.dev-reference-%2300ADD8?style=for-the-badge&logo=go&logoColor=white)](https://pkg.go.dev/github.com/bgreenwell/git-ego)
[![Latest Release](https://img.shields.io/github/v/release/bgreenwell/git-ego?style=for-the-badge&color=%23107C41)](https://github.com/bgreenwell/git-ego/releases/latest)

[![License: MIT](https://img.shields.io/badge/License-MIT-%232196F3.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/go-1.24%2B-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Easy Install](https://img.shields.io/badge/Easy%20Install-Homebrew%20%7C%20Scoop%20%7C%20WinGet-%23FBB040?style=for-the-badge)](#installation)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=for-the-badge)](https://github.com/bgreenwell/git-ego/releases/latest)

> Git identity manager with directory-based profile switching.

`git-ego` manages separate Git identities for work, personal projects, and clients. Profiles can set `user.name`, `user.email`, SSH keys, signing keys, and HTTPS credentials.

Use either `git-ego <command>` or `git ego <command>`. When `git-ego` is on
`PATH`, Git resolves `git ego` to the same executable. This README uses
`git ego` for user-facing examples; Git's credential-helper configuration must
continue to invoke `git-ego credential` directly.

---

## Key features

- Directory-based profile switching using Git `includeIf`.
- Per-profile commit identity, SSH configuration, signing keys, and HTTPS host scopes.
- PATs stored in the operating-system keychain and served through Git's credential-helper protocol.
- Optional pre-commit identity check.
- Fail-closed credential selection verified against Git's effective identity.

## Installation

### Package managers

**macOS / Linux (Homebrew):**

```bash
brew install bgreenwell/tap/git-ego
```

**Windows (Scoop):**

```powershell
scoop bucket add bgreenwell https://github.com/bgreenwell/scoop-bucket
scoop install git-ego
```

**Windows (WinGet):** _(Coming soon — pending initial PR merge)_

```powershell
winget install --id bgreenwell.GitEgo
```

**Nix (Linux / macOS):**

```bash
nix profile install github:bgreenwell/git-ego
```

### Go

Or install from source with [Go](https://go.dev/dl/) 1.24 or newer:

```bash
go install github.com/bgreenwell/git-ego@latest
```

Ensure your Go bin directory (typically `~/go/bin`) is in `PATH` when using
`go install`.

### GitHub account-switch smoke test

From a source checkout, run the included account-switch smoke test against a
repository you control. It uses the test-account PAT already stored in the
keychain (or stores `GH_GIT_EGO` when supplied), verifies retrieval and host
scoping, writes a timestamped entry to a smoke-test changelog, pushes through
the test profile, and restores your regular profile.

```bash
GIT_EGO_SMOKE_RETURN_PROFILE=personal \
  scripts/gh-account-switch-smoke-test.sh --push
```

## One-time setup: Configure Git

Configure GitHub to use `git-ego` as its credential helper. This leaves other
hosts and their existing helpers untouched.

```bash
# Use git-ego only for github.com. The empty value resets helpers for this URL.
git config --global credential.https://github.com.helper ""
git config --global --add credential.https://github.com.helper "!git-ego credential"
```

## Example usage

Here’s a typical workflow for setting up and using `git-ego`.

#### 1\. Add your profiles

First, define your different identities.

```bash
# A simple personal profile
git ego add personal --name "Brandon" --email "brandon.personal@email.com" --username "bgreenwell-personal"

# A work profile that uses a specific SSH key
git ego add work-ssh --name "Brandon Greenwell" --email "brandon.work@company.com" --username "bgreenwell-work" --ssh-key ~/.ssh/id_work

# A client profile that uses a PAT for HTTPS. Store the token separately so
# it never enters shell history or the process list.
git ego add client-abc --name "Brandon G." --email "brandon@client-abc.com" --username "bgreenwell-client"
printf '%s' "$CLIENT_GITHUB_TOKEN" | git ego pat set client-abc

# A non-GitHub profile must explicitly name its HTTPS host.
git ego add gitlab-work --name "Brandon" --email "brandon@company.com" --username "bgreenwell" --host gitlab.company.com
```

#### 2\. List your configured profiles

Use the `list` (or `ls`) command to see all the profiles you’ve saved.

```bash
git ego list --check-pats
```

```
ACTIVE  PROFILE     NAME                 EMAIL                       ATTRIBUTES
------  -------     ----                 -----                       ----------
        client-abc  Brandon G.           brandon@client-abc.com      [PAT]
        personal    Brandon              brandon.personal@email.com
        work-ssh    Brandon Greenwell    brandon.work@company.com    [SSH]
```

#### 3\. Set a global default

The `use` command sets your default identity for repositories without a more specific rule. It updates authoritative YAML and reconciles one managed include at the end of your global `.gitconfig`.

```bash
git ego use personal
```

```
✓ Set active profile to 'personal'.
```

`git-ego list` marks the active profile with an asterisk:

```
ACTIVE  PROFILE     NAME                 EMAIL                       ATTRIBUTES
------  -------     ----                 -----                       ----------
        client-abc  Brandon G.           brandon@client-abc.com      [PAT]
*       personal    Brandon              brandon.personal@email.com
        work-ssh    Brandon Greenwell    brandon.work@company.com    [SSH]
```

#### 4\. Configure automatic switching

Now, tell `git-ego` which profiles to use for which project directories.

```bash
git ego auto ~/dev/work/ work-ssh
git ego auto ~/dev/personal/ personal

# A normalized path can have only one profile. Change it explicitly:
git ego auto ~/dev/personal/ work-ssh --replace

# Review or remove one rule later.
git ego auto list
git ego auto rm ~/dev/personal/
```

When you `cd` into `~/dev/work/any-repo`, your `user.name`, `user.email`, and `sshCommand` will be automatically switched to the `work-ssh` profile.

For a repository-specific safety expectation, create a `.gitego` file at the repository root containing one profile name (for example `work-ssh`). This is an assertion only: it never changes identity or selects a PAT. Select a repository-local identity with `git ego use work-ssh --local`. An invalid, unreadable, or mismatched assertion makes the credential helper refuse credentials and makes the commit hook abort. Add `.gitego` to the repository's local `.gitignore` if it should not be shared.

### Upgrading to v0.3

Run `git ego doctor --repair` once. It assigns stable internal credential IDs, migrates legacy name-keyed PATs, removes old git-ego include rules, and regenerates the managed configuration. A validated legacy fallback keeps existing installations usable until repair, but it is temporary and will be removed in a later major release.

-----

## Commands

| Command | Alias | Description |
|---|---|---|
| `git ego add <name>` | | Adds a new user profile. |
| `git ego rm <name>` | `remove` | Removes a saved user profile, asking for confirmation. |
| `git ego list` | `ls` | Lists all saved user profiles and their attributes. |
| `git ego list --check-pats` | | Opts in to keyring checks and PAT markers. |
| `git ego use <name>` | | Sets a profile as the active global default. |
| `git ego use <name> --local` | | Applies an identity only in the current repository. |
| `git ego auto <path> <name>` | | Sets a profile to be used automatically for a directory. |
| `git ego auto <path> <name> --replace` | | Reassigns an existing normalized directory rule. |
| `git ego auto list` | | Lists configured auto-switch rules. |
| `git ego auto rm <path>` | | Removes one auto-switch rule. |
| `git ego status` | | Displays the current effective Git user and its source. |
| `git ego edit <name>` | | Edits an existing user profile's attributes. |
| `git ego pat set <name>` | | Stores a PAT read from standard input. |
| `git ego pat delete <name>` | | Deletes a PAT from the secure keyring. |
| `git ego install-hook` | | Installs a pre-commit hook in the current repo. |
| `git ego completion <shell>` | | Generates shell completion scripts. |
| `git ego doctor` | | Checks markers, generated Git files, credential IDs, rules, and helper configuration. |
| `git ego doctor --repair` | | Migrates legacy credentials and regenerates managed Git files. |
| `git ego profiles export <file>` | | Exports profiles and auto-rules without PATs. |
| `git ego profiles import <file>` | | Imports profiles and auto-rules without PATs. |
| `git ego profiles import <file> --replace` | | Replaces non-empty state and detaches its PAT associations. |
| `git ego --version` | `-v` | Prints the application version. |

## How it works

`git-ego` uses Git conditional includes and the credential-helper protocol.

### Visualizing the workflow

```mermaid
graph LR
    A[Enter repository] --> B[Git evaluates managed includes];
    B --> C[Effective identity and gitego.profile];
    C --> D{Explicit expectation?};
    D -- .gitego or auto-rule --> E[Verify expected profile];
    D -- None --> F[Verify effective profile identity];
    E --> F;
    F --> G{Consistent?};
    G -- No --> H[Refuse credentials or abort commit];
    G -- Yes --> I{HTTPS operation?};
    I -- Yes --> J[Read credential-ID PAT from keychain];
    J --> K[Return host-scoped credential];
    I -- No --> L[Use profile SSH configuration];
```

### 1\. Identity switching: `includeIf`

For managing your commit identity (`user.name`, `user.email`) and SSH keys, `git-ego` uses a Git feature called conditional includes.

The generated layout is:

1. `~/.gitego/profiles/<name>.gitconfig` contains identity, SSH/signing settings, and `gitego.profile`.
2. `~/.gitego/default.gitconfig` includes the active global profile.
3. `~/.gitego/includes.gitconfig` includes the default first, followed by directory rules from broad to specific.
4. One include at the end of `~/.gitconfig` loads `includes.gitconfig`.

Git evaluates all matching rules in order, so a nested rule can safely
override a broader directory rule.

### 2\. Authentication: credential helper

For handling Personal Access Tokens (PATs) with HTTPS remotes, `git-ego` acts as a **Git credential helper**.

1.  **Configuration**: The host-scoped commands in the setup section tell Git to invoke `git-ego credential` for `https://github.com` without replacing helpers for unrelated hosts.
2.  **Execution**: When you run `git push` or `git pull` on an HTTPS remote, Git executes `git-ego credential` in the background and pipes it the protocol, host, and path information.
3.  **Context Resolution**: Git's effective `gitego.profile` marker is the credential source of truth. The helper verifies its effective `user.name` and `user.email`, then checks any repository-root `.gitego` assertion or most-specific auto-rule.
4.  **Secure Retrieval**: It then fetches the corresponding PAT for that profile from your operating system's native, secure keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service).
5.  **Response**: Finally, it prints the username and PAT to standard output, which Git reads to complete the authentication.

### Security model

PAT handling:

  * **No Plaintext PATs:** Personal Access Tokens (PATs) are never stored in plaintext in the configuration file.
  * **Secure OS Keychain:** `git-ego` uses the native, secure keychain of your operating system to store and retrieve your PATs. This is the same secure storage that tools like Docker and other credential managers use.
  * **Host Scoping:** The credential helper only provides a token for an HTTPS host explicitly configured on the selected profile. It ignores requests for other hosts and all `store` and `erase` operations.
  * **Consistency Checks:** A missing/unknown marker, identity mismatch, or explicit expectation mismatch produces no credential protocol output.
  * **In-Memory Use:** The token is passed directly to Git in memory and is not logged or stored elsewhere.

### Persistence and repair

`~/.gitego/config.yaml` is the authoritative desired state. Mutating commands
serialize access with `~/.gitego/config.lock`, atomically save validated YAML,
and then regenerate the managed Git files. If regeneration is interrupted or
fails, the YAML remains valid and `git ego doctor --repair` can rebuild the
derived files.

Profiles use an internal random credential ID for keychain accounts so renames,
imports, and removals cannot accidentally reuse a stale name-keyed PAT.
Portable exports omit these IDs, and imports always generate fresh ones.


## Contributing

Contributions are welcome\! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License.
