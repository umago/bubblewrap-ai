package main

import (
	_ "embed"
	"encoding/json"
	"os"
)

//go:embed defaults.json
var defaultConfigJSON []byte

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
	var cfg Config
	if err := json.Unmarshal(defaultConfigJSON, &cfg); err != nil {
		panic("bwai: invalid defaults.json: " + err.Error())
	}
	return cfg
}

// loadConfig reads the config file at the given path if it exists and returns the resulting Config.
// Fields omitted from the file fall back to the defaults.
func loadConfig(path string) (cfg Config, err error) {
	cfg = defaultConfig()
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
