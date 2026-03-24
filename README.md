# bubblewrap-ai

Runs AI coding agents (Claude, Gemini, Goose) inside a [bubblewrap](https://github.com/containers/bubblewrap) sandbox. The host filesystem is read-only; only the current project directory and the dotfiles you whitelist are accessible.

## Requirements

- Linux
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

## Configuration

`bwai` works out of the box with no config file. To customise behaviour, create `~/.bwai.json`:

```json
{
  "bwrap_path": "/usr/bin/bwrap",
  "command": ["bash"],
  "bwrap_extra_args": [""],
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
    ".docker",
    ".password-store",
    ".bashrc",
    ".bashrc.d",
    ".bash_history*",
    ".netrc",
    ".npmrc",
    ".pypirc",
    ".config/Bitwarden",
    ".config/gh",
    ".config/gcloud",
    ".config/op",
    ".config/helm",
    ".config/git"
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

`home_blocked` takes precedence over `home_allowed`.
