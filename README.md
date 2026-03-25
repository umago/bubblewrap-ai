# bubblewrap-ai

Runs AI coding agents (Claude, Gemini, Goose) inside a [bubblewrap](https://github.com/containers/bubblewrap) sandbox. The host filesystem is read-only; only the current project directory and the dotfiles you whitelist are accessible.

## Requirements

- Linux (Wayland)
- [`bwrap`](https://github.com/containers/bubblewrap) installed (e.g. `sudo dnf install bubblewrap` or `sudo apt install bubblewrap`)

## Install

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

To skip the shell and launch an agent directly, set the `command` field in `~/.bwai.json`:

```json
{ "command": ["claude"] }
```

## Configuration

`bwai` works out of the box with no config file. To customise behaviour, create `~/.bwai.json`.

To see the full default configuration as a starting point, run:

```sh
bwai --dump-config > ~/.bwai.json
```

Example `~/.bwai.json`:

```json
{
  "bwrap_path": "bwrap",
  "command": ["bash"],
  "bwrap_extra_args": [],
  "home_allowed": [
    ".claude",
    ".gemini",
    ".claude.json",
    ".config/goose",
    ".local/state",
    ".local/share/goose",
    ".cache"
  ],
  "home_blocked": [
    ".gnupg",
    ".ssh",
    ".pki",
    ".aws",
    ".kube",
    ".azure",
    ".password-store",
    ".bashrc",
    ".bashrc.d",
    ".bash_history*",
    ".config/Bitwarden"
  ]
}
```

| Field | Description | Default |
|---|---|---|
| `bwrap_path` | Path to the `bwrap` binary | `"bwrap"` |
| `command` | Command (and args) to run inside the sandbox | `["bash"]` |
| `bwrap_extra_args` | Extra arguments forwarded to `bwrap` (e.g. `--unshare-net`) | `[]` |
| `home_allowed` | Dotfiles/dirs in `$HOME` the agent may read and write | see above |
| `home_blocked` | Dotfiles/dirs in `$HOME` that are never exposed | see above |

`home_allowed` takes precedence over `home_blocked`.
