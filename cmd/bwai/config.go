package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	// Path to the bwrap binary. Defaults to "bwrap"
	BwrapPath string `json:"bwrap_path"`

	// Extra arguments passed to bwrap. Use this to add --unshare-net, --setenv HTTP_PROXY, etc...
	BwrapExtraArgs []string `json:"bwrap_extra_args"`

	// Default command to run. Defaults to ["bash"]
	Command []string `json:"command"`

	// Files and directories in $HOME that agents need write access to
	HomeAllow []string `json:"home_allow"`

	// Sensitive files and directories in $HOME that must never be exposed
	HomeBlock []string `json:"home_block"`

	// Environment variables from the host that are passed into the sandbox
	EnvAllow []string `json:"env_allow"`
}

func defaultConfig() Config {
	return Config{
		BwrapPath:      "bwrap",
		BwrapExtraArgs: []string{"--unshare-pid", "--unshare-ipc"},
		Command:        []string{"bash"},
		EnvAllow: []string{
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
			// Claude
			"ANTHROPIC_API_KEY",
			// Claude model selection / pinning
			"ANTHROPIC_MODEL",
			"ANTHROPIC_DEFAULT_OPUS_MODEL",
			"ANTHROPIC_DEFAULT_SONNET_MODEL",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL",
			// Claude Code on Google Vertex AI
			"CLAUDE_CODE_USE_VERTEX",
			"CLOUD_ML_REGION",
			"ANTHROPIC_VERTEX_PROJECT_ID",
			// Gemini / Google
			"GEMINI_API_KEY",
			"GOOGLE_API_KEY",
			"GCLOUD_PROJECT",
			"GOOGLE_CLOUD_PROJECT",
			// Goose (uses provider keys above + its own config)
			"GOOSE_PROVIDER",
			"GOOSE_MODEL",
			// OpenAI-compatible providers (used by Goose and others)
			"OPENAI_API_KEY",
			"OPENAI_API_BASE",
			// OpenRouter
			"OPENROUTER_API_KEY",
		},
		HomeAllow: []string{
			".claude",
			".gemini",
			".claude.json",
			".config/goose",
			".config/gcloud",
			".local/state",
			".local/share/goose",
			".cache",
		},
		HomeBlock: []string{
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
			".config/Bitwarden",
		},
	}
}

// loadConfig reads ~/.bwai.json if it exists and returns the resulting Config.
// Fields omitted from the file fall back to the defaults
func loadConfig(home string) (cfg Config, err error) {
	cfg = defaultConfig()
	path := filepath.Join(home, ".bwai.json")
	var f *os.File
	f, err = os.Open(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Package-level vars set in main()
var homeAllow []string
var homeBlock []string
