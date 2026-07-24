# AGENTS.md

## What This Project Is

`bwai` is a CLI tool that runs AI coding agents (Claude, Gemini, Goose) inside a [bubblewrap](https://github.com/containers/bubblewrap) sandbox. The sandbox enforces a read-only host filesystem, whitelisted dotfiles and environment variables, and restricts writes to the current working directory only — preventing agents from accessing sensitive data like SSH keys, AWS credentials, or GPG keys.

## Project Layout

```
cmd/bwai/
  main.go       # Entry point, flag parsing, bwrap command construction
  config.go     # Config struct and JSON loading (~/.bwai.json)
  defaults.json # Built-in default configuration (embedded into the binary at build time)
  mounts.go     # Filesystem mount logic (home, DNS, GPU, shm)
  bwrap.go      # Low-level bwrap argument helpers (roBind, rwBind, devBind, tmpfs)
  update.go     # Self-update from GitHub releases with SHA-256 verification
  version.go    # Version constant (set at build time via ldflags)
  *_test.go     # Unit tests for each module
scripts/
  check.sh      # Runs fmt, lint, and tests (used in CI and pre-commit)
Makefile        # build, install, test, fmt, lint, clean targets
```

## Language and Dependencies

- Pure Go 1.21, zero external dependencies (stdlib only)
- Requires `bwrap` installed on the host system (Linux only)

## Build and Test

```sh
make build    # Compiles to bin/bwai
make test     # Runs all unit tests
make lint     # Runs golangci-lint
make fmt      # Formats code
```

CI runs `scripts/check.sh` which chains fmt-check, lint, and test.

## Configuration

The config file lives at `~/.bwai.json`. Key fields:

| Field | Purpose |
|---|---|
| `bwrap_path` | Path to the `bwrap` binary |
| `bwrap_extra_args` | Extra flags passed directly to bwrap. Each element is split on whitespace, so `"--ro-bind /var /var"` and `"--ro-bind", "/var", "/var"` are equivalent |
| `command` | Default command to run inside the sandbox |
| `home_allow` | Dotfiles/paths from `$HOME` to expose (read-only) |
| `home_block` | Dotfiles/paths to explicitly block |
| `env_allow` | Environment variables to pass through |

`home_allow` takes precedence over `home_block`. Patterns support glob suffixes (e.g., `.bash_history*`) and nested paths (e.g., `.config/goose`).

## Key Design Decisions

- **Read-only by default**: The entire host OS tree is bind-mounted read-only. Only the current working directory is writable.
- **No parent-process exit issue**: Non-bash commands are wrapped in `bash -i -c` to avoid agent processes becoming orphaned when the shell exits.
- **GPU support**: `gpuMounts()` in `mounts.go` auto-detects and mounts NVIDIA and DRI devices.
- **Safe self-update**: Downloads binary, verifies SHA-256 digest, replaces atomically with rollback on failure.
- **Zero dependencies**: Keeps the supply chain minimal; everything uses Go stdlib.

## CLI Flags

```
bwai                        # Open sandboxed bash shell
bwai -c claude              # Launch claude directly inside sandbox
bwai --ro-dir /some/path    # Expose extra read-only directory
bwai -- --some-agent-flag   # Pass args to the configured command
bwai update                 # Self-update to latest GitHub release
bwai --dump-config          # Print default config as JSON
```

## Adding New Features

- **New mount type**: Add a helper function in `mounts.go` following the existing pattern (`[]string` return, appended in `main.go`).
- **New CLI flag**: Add to `main.go` using the `flag` package; wire into the bwrap args slice.
- **New config field**: Add to the `Config` struct in `config.go` and add the default value to `defaults.json`.
- **New agent support**: Add its dotfiles to `home_allow` and its API key to `env_allow` in `defaults.json`.

## Testing

Tests live alongside source files (`*_test.go`). They use only `testing` from stdlib. Run with `make test`. There is no integration test harness — unit tests cover mount logic, bwrap arg construction, and the update mechanism.
