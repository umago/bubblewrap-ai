package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
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

	cfg, err := loadConfig(home)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bwai: warning: could not load %s: %v\n", filepath.Join(home, ".bwai.json"), err)
	}
	homeAllowed = cfg.HomeAllowed
	homeBlocked = cfg.HomeBlocked

	fmt.Printf("bwai: sandboxed in %s\n", currentDir)

	command := cfg.Command
	args := []string{
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
	}
	args = append(args, shmMount()...)
	args = append(args,
		// Virtual filesystems
		"--proc", "/proc",
		"--tmpfs", "/tmp",
		"--tmpfs", "/run",
	)
	args = append(args, dnsMounts()...)
	// Wayland display
	args = append(args, displayArgs()...)
	// Home directory
	args = append(args, tmpfs(home)...)
	args = append(args, homeMounts(home)...)
	args = append(args,
		// Current directory
		"--bind", currentDir, currentDir,
		"--chdir", currentDir,
		// Namespace isolation
		"--die-with-parent",
		"--unshare-pid",
		"--unshare-ipc",
	)
	args = append(args, cfg.BwrapExtraArgs...)

	// Inject a minimal ~/.bashrc into the sandbox so the shell prompt looks nice
	// Pipe the content directly via --file <fd> <dest> to avoid any temp files
	var extraFiles []*os.File
	bashrcR, bashrcW, pipeErr := os.Pipe()
	if pipeErr == nil {
		fmt.Fprint(bashrcW, "PS1='[🫧] > '\n")
		bashrcW.Close()
		// ExtraFiles[0] becomes fd 3 (after stdin/stdout/stderr)
		extraFiles = append(extraFiles, bashrcR)
		args = append(args, "--file", "3", filepath.Join(home, ".bashrc"))
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
