package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Tells whether name matches any direct (slash-free) pattern in the list
func matchesDirect(patterns []string, name string) bool {
	for _, pattern := range patterns {
		if strings.Contains(pattern, "/") {
			continue
		}
		if matched, err := filepath.Match(pattern, name); err == nil && matched {
			return true
		}
	}
	return false
}

// Calls mount() for each sub-path entry in patterns that exists on disk
func subPathMounts(home string, patterns []string, mount func(string) []string) []string {
	var args []string
	for _, pattern := range patterns {
		if !strings.Contains(pattern, "/") {
			continue
		}
		p := filepath.Join(home, pattern)
		if _, err := os.Stat(p); err == nil {
			args = append(args, mount(p)...)
		}
	}
	return args
}

// Mount every dotfile and dotdir in $HOME, except those that are sensitive.
// Entries in homeAllowed are mounted as read-write. Everything else are
// mounted as read-only or hidden with tmpfs
func homeMounts(home string) []string {
	entries, err := os.ReadDir(home) // already sorted by name
	if err != nil {
		return nil
	}
	var args []string
	for _, entry := range entries {
		name := entry.Name()
		if matchesDirect(homeBlocked, name) || name[0] != '.' {
			continue
		}
		p := filepath.Join(home, name)
		if matchesDirect(homeAllowed, name) {
			args = append(args, rwBind(p)...)
		} else {
			args = append(args, roBind(p)...)
		}
	}
	// Apply sub-path overrides after all parent dirs are mounted.
	// Blocked sub-paths must be hidden first, then allowed sub-paths can
	// selectively re-expose specific files within blocked directories
	args = append(args, subPathMounts(home, homeBlocked, func(p string) []string { return tmpfs(p) })...)
	args = append(args, subPathMounts(home, homeAllowed, func(p string) []string { return rwBind(p) })...)
	return args
}

// Bind the Wayland socket and pass the display env vars into the sandbox
func displayArgs() []string {
	xdgRuntime := os.Getenv("XDG_RUNTIME_DIR")
	waylandDisp := os.Getenv("WAYLAND_DISPLAY")
	args := rwBind(xdgRuntime)
	args = append(args,
		"--setenv", "XDG_RUNTIME_DIR", xdgRuntime,
		"--setenv", "WAYLAND_DISPLAY", waylandDisp,
	)
	return args
}

// Bind /dev/shm (shared memory) if it exists on this host
func shmMount() []string {
	p := "/dev/shm"
	if info, err := os.Stat(p); err == nil && info.IsDir() {
		return devBind(p)
	}
	return nil
}

// Restore /run/systemd/resolve after the --tmpfs /run overlay,
// otherwise the sandboxed agent won't be able to read resolv.conf and
// won't have network access
func dnsMounts() []string {
	p := "/run/systemd/resolve"
	if info, err := os.Stat(p); err == nil && info.IsDir() {
		return roBind(p)
	}
	return nil
}
