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
	HomeAllowed []string `json:"home_allowed"`

	// Sensitive files and directories in $HOME that must never be exposed
	HomeBlocked []string `json:"home_blocked"`
}

func defaultConfig() Config {
	return Config{
		BwrapPath: "bwrap",
		Command:   []string{"bash"},
		HomeAllowed: []string{
			".claude",
			".gemini",
			".claude.json",
			".config/goose",
			".local/state",
			".local/share/goose",
			".cache",
		},
		HomeBlocked: []string{
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
			".config/git",
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
var homeAllowed []string
var homeBlocked []string
