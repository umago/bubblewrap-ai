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
bwai -c "claude --model gemini-2.0-flash-exp"
```

## Configuration

`bwai` works out of the box with no config file. To customise behaviour, create `~/.bwai.json` as a global config. This can be overridden per-run with the `--config` flag:

```sh
bwai --config /path/to/my-config.json
```

To see the full default configuration as a starting point, run:

```sh
bwai --dump-config > ~/.bwai.json
```

Example `~/.bwai.json`:

```json
{
  "bwrap_path": "bwrap",
  "bwrap_extra_args": ["--unshare-pid", "--unshare-ipc"],
  "command": ["bash"],
  "home_allow": [
    ".claude",
    ".gemini",
    ".claude.json",
    ".config/goose",
    ".config/gcloud",
    ".local/state",
    ".local/share/goose",
    ".cache"
  ],
  "home_block": [
    ".gnupg",
    ".ssh",
    ".pki",
    ".aws",
    ".kube",
    ".azure",
    ".bashrc",
    ".bashrc.d",
    ".password-store",
    ".bash_history*",
    ".config/Bitwarden"
  ],
  "env_allow": [
    "TERM",
    "COLORTERM",
    "LANG",
    "LC_ALL",
    "LC_MESSAGES",
    "LC_CTYPE",
    "HOME",
    "USER",
    "LOGNAME",
    "PATH",
    "ANTHROPIC_API_KEY",
    "ANTHROPIC_MODEL",
    "ANTHROPIC_DEFAULT_OPUS_MODEL",
    "ANTHROPIC_DEFAULT_SONNET_MODEL",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL",
    "CLAUDE_CODE_USE_VERTEX",
    "CLOUD_ML_REGION",
    "ANTHROPIC_VERTEX_PROJECT_ID",
    "GEMINI_API_KEY",
    "GOOGLE_API_KEY",
    "GCLOUD_PROJECT",
    "GOOGLE_CLOUD_PROJECT",
    "GOOSE_PROVIDER",
    "GOOSE_MODEL",
    "OPENAI_API_KEY",
    "OPENAI_API_BASE",
    "OPENROUTER_API_KEY",
  ]
}
```

| Field | Description | Default |
|---|---|---|
| `bwrap_path` | Path to the `bwrap` binary | `"bwrap"` |
| `bwrap_extra_args` | Extra arguments forwarded to `bwrap` (e.g. `--unshare-net`) | `["--unshare-pid", "--unshare-ipc"]` |
| `command` | Command (and args) to run inside the sandbox | `["bash"]` |
| `home_allow` | Dotfiles/dirs in `$HOME` the agent may read and write | see above |
| `home_block` | Dotfiles/dirs in `$HOME` that are never exposed | see above |
| `env_allow` | Environment variables from the host passed into the sandbox | see above |

`home_allow` takes precedence over `home_block`.
