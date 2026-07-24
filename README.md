# bubblewrap-ai

Runs AI coding agents (Claude, Gemini, Goose) inside a [bubblewrap](https://github.com/containers/bubblewrap) sandbox. The host filesystem is read-only, only the current project directory and the dotfiles you whitelist are accessible. The sandbox also starts with a clean environment, only variables explicitly allowed are visible to the agent.

## Requirements

- Linux
- [`bwrap`](https://github.com/containers/bubblewrap) installed (e.g. `sudo dnf install bubblewrap` or `sudo apt install bubblewrap`)

## Install

### From GitHub Releases (recommended)

```sh
curl -Lo ~/.local/bin/bwai https://github.com/umago/bubblewrap-ai/releases/latest/download/bwai
chmod +x ~/.local/bin/bwai
```

### From source

```sh
make build
cp bin/bwai ~/.local/bin/
```

## Update

To update `bwai` to the latest release:

```sh
bwai update
```

This downloads the latest binary from [GitHub releases](https://github.com/umago/bubblewrap-ai/releases), verifies its SHA-256 digest, and replaces the running binary in-place.

## Usage

Run `bwai` from inside the project directory you want to give the agent access to:

```sh
cd ~/my-project
bwai
```

By default, `bwai` opens a sandboxed `bash` shell. From there you can launch any agent:

```sh
[🫧] > claude
[🫧] > goose
[🫧] > gemini
```

### Running a command directly

To skip the shell and launch an agent (or any command) directly, you can either:

1. Set the `command` field in `~/.bwai.json`:

```json
{ "command": ["claude"] }
```

2. Use the `--command` (or `-c`) CLI flag, which overrides the config file:

```sh
bwai --command claude

```

To append arguments to the command configured in `~/.bwai.json`, use `--`:

```sh
# With "command": ["goose"] in config
bwai -- session -r  # runs "goose session -r" to resume a session

# With "command": ["claude"] in config
bwai -- --model gemini-2.0-flash-exp  # runs "claude --model gemini-2.0-flash-exp"
```

Everything after `--` is passed as extra arguments to the resolved command.

### Exposing extra directories (read-only)

Use `--ro-dir` to give the agent read-only access to directories outside the current project. This is useful when the project you are working on depends on another local project that you want the agent to use as reference:

```sh
bwai --ro-dir ../other-project
bwai --ro-dir /absolute/path/to/lib --ro-dir /another/path
```

The flag is repeatable and accepts both relative and absolute paths. Each directory is mounted at the same absolute path inside the sandbox. `bwai` will refuse to start if any of the given paths does not exist.

## Configuration

`bwai` works out of the box with no config file. To customise behaviour, create `~/.bwai.json` as a global config. To use a different file for a single run:

```sh
bwai --config /path/to/my-config.json
```

To use the built-in defaults as a starting point, run `bwai --dump-config > ~/.bwai.json`, or browse them at [`cmd/bwai/defaults.json`](cmd/bwai/defaults.json).

The available fields are:

| Field | Description |
|---|---|
| `bwrap_path` | Path to the `bwrap` binary |
| `bwrap_extra_args` | Extra arguments forwarded to `bwrap` (e.g. `--unshare-net`). Each element is split on whitespace, so `"--ro-bind /var /var"` and `"--ro-bind", "/var", "/var"` are equivalent |
| `command` | Command (and args) to run inside the sandbox |
| `home_allow` | Dotfiles/dirs in `$HOME` the agent may read and write |
| `home_block` | Dotfiles/dirs in `$HOME` that are never exposed |
| `env_allow` | Environment variables from the host passed into the sandbox |

`home_block` takes precedence over `home_allow` at the same nesting level. However, the two can be combined at different nesting levels to achieve more granular control:

- **Block a sub-path inside an allowed directory** — `home_allow: [".cache"]` + `home_block: [".cache/sccache"]` exposes `.cache` (read-write) but hides `.cache/sccache` inside it.
- **Re-expose a sub-path inside a blocked directory** — `home_block: [".cache"]` + `home_allow: [".cache/sccache"]` hides everything in `.cache` except `.cache/sccache`, which is available read-write.
