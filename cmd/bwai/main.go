package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	dumpConfig := flag.Bool("dump-config", false, "Print the default configuration JSON and exit")
	configFlag := flag.String("config", "", "Path to a config file (overrides ~/.bwai.json)")
	commandFlag := flag.String("command", "", "Command to run inside the sandbox (overrides config and default)")
	flag.StringVar(commandFlag, "c", "", "Shorthand for --command")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	if *dumpConfig {
		cfg := defaultConfig()
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "bwai: failed to encode config: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bwai: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bwai: cannot determine current directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(home, ".bwai.json")
	if *configFlag != "" {
		configPath = *configFlag
	}
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bwai: warning: could not load %s: %v\n", configPath, err)
	}
	homeAllow = cfg.HomeAllow
	homeBlock = cfg.HomeBlock

	command := cfg.Command
	if *commandFlag != "" {
		command = []string{"bash", "-i", "-c", *commandFlag}
	}

	fmt.Printf("bwai: sandboxed in %s\n", currentDir)
	args := []string{
		// Clear the inherited environment; only whitelisted vars are passed through below
		"--clearenv",
	}
	for _, key := range cfg.EnvAllow {
		if val, ok := os.LookupEnv(key); ok {
			args = append(args, "--setenv", key, val)
		}
	}
	args = append(args,
		// Read-only OS tree
		"--ro-bind", "/usr", "/usr",
		"--ro-bind", "/etc", "/etc",
		"--ro-bind", "/bin", "/bin",
		"--ro-bind", "/lib", "/lib",
		"--ro-bind", "/lib64", "/lib64",
		"--ro-bind", "/opt", "/opt",
		"--ro-bind", "/sys", "/sys",
		// Device nodes
		"--dev", "/dev",
	)
	args = append(args, shmMount()...)
	args = append(args,
		// Virtual filesystems
		"--proc", "/proc",
		"--tmpfs", "/tmp",
		"--tmpfs", "/run",
	)
	args = append(args, dnsMounts()...)
	// Home directory
	args = append(args, tmpfs(home)...)
	args = append(args, homeMounts(home)...)
	args = append(args,
		// Current directory
		"--bind", currentDir, currentDir,
		"--chdir", currentDir,
		// Namespace isolation
		"--die-with-parent",
	)
	args = append(args, cfg.BwrapExtraArgs...)

	// Inject a minimal rcfile so PS1 is set after /etc/bashrc runs, without
	// creating any file at ~/.bashrc (which is blocked). Write to /tmp/bwai.sh
	// (inside the --tmpfs /tmp) and point bash at it via --rcfile
	var extraFiles []*os.File
	if len(command) > 0 && filepath.Base(command[0]) == "bash" {
		bashrcR, bashrcW, pipeErr := os.Pipe()
		if pipeErr == nil {
			_, _ = fmt.Fprint(bashrcW, "PS1='[🫧] > '\n")
			_ = bashrcW.Close()
			// ExtraFiles[0] becomes fd 3 (after stdin/stdout/stderr)
			extraFiles = append(extraFiles, bashrcR)
			args = append(args, "--file", "3", "/tmp/bwai.sh")
			command = append([]string{command[0], "--rcfile", "/tmp/bwai.sh"}, command[1:]...)
		}
	}

	args = append(args, command...)

	// Execute the bubblewrap command
	cmd := exec.Command(cfg.BwrapPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = extraFiles

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "bwai: %v\n", err)
		os.Exit(1)
	}
}
