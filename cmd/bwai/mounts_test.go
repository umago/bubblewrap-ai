package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesDirect(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		input    string
		want     bool
	}{
		{"exact match", []string{".ssh"}, ".ssh", true},
		{"no match", []string{".ssh"}, ".gnupg", false},
		{"glob asterisk suffix", []string{".bash_history*"}, ".bash_history", true},
		{"glob asterisk matches extension", []string{".bash_history*"}, ".bash_history.bak", true},
		{"slash pattern is skipped", []string{".config/goose"}, ".config", false},
		{"slash pattern does not match leaf", []string{".config/goose"}, "goose", false},
		{"empty patterns", []string{}, ".ssh", false},
		{"first of multiple matches", []string{".ssh", ".gnupg"}, ".ssh", true},
		{"second of multiple matches", []string{".ssh", ".gnupg"}, ".gnupg", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesDirect(tt.patterns, tt.input)
			if got != tt.want {
				t.Errorf("matchesDirect(%v, %q) = %v, want %v", tt.patterns, tt.input, got, tt.want)
			}
		})
	}
}

func TestSubPathMounts(t *testing.T) {
	home := t.TempDir()

	existingDir := filepath.Join(home, ".config", "goose")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatal(err)
	}

	patterns := []string{
		".ssh",            // direct (no slash) must be skipped
		".config/goose",   // sub-path, exists on disk
		".config/missing", // sub-path, does not exist
	}

	var mounted []string
	mount := func(p string) []string {
		mounted = append(mounted, p)
		return []string{"--test", p}
	}

	args := subPathMounts(home, patterns, mount)

	if len(mounted) != 1 || mounted[0] != existingDir {
		t.Errorf("expected mount called once with %q, got %v", existingDir, mounted)
	}
	if !containsSequence(args, "--test", existingDir) {
		t.Errorf("args missing expected sequence; got %v", args)
	}
}

func TestHomeMounts(t *testing.T) {
	home := t.TempDir()

	// Mock package-level globals for the duration of the test.
	origAllowed := homeAllow
	origBlocked := homeBlock
	t.Cleanup(func() {
		homeAllow = origAllowed
		homeBlock = origBlocked
	})

	homeAllow = []string{".claude", ".config/goose"}
	homeBlock = []string{".ssh", ".config/secret"}

	mkDir := func(rel string) string {
		p := filepath.Join(home, rel)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	mkFile := func(rel string) string {
		p := filepath.Join(home, rel)
		if err := os.WriteFile(p, nil, 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	claudeDir := mkDir(".claude")        // allowed directory: --bind
	_ = mkDir(".ssh")                    // blocked directory: not mounted
	vimDir := mkDir(".vim")              // unclassified dotdir: --ro-bind
	_ = mkFile("README.md")              // non-dotfile: not mounted
	gooseDir := mkDir(".config/goose")   // allowed sub-path: --bind (last)
	secretDir := mkDir(".config/secret") // blocked sub-path: --tmpfs (before allowed)

	args := homeMounts(home)

	t.Run("allowed dotdir is rw-bound", func(t *testing.T) {
		if !containsSequence(args, "--bind", claudeDir, claudeDir) {
			t.Errorf("expected --bind for allowed dir %q; args: %v", claudeDir, args)
		}
	})

	t.Run("blocked dotdir is not mounted at parent level", func(t *testing.T) {
		sshDir := filepath.Join(home, ".ssh")
		if containsSequence(args, "--ro-bind", sshDir, sshDir) || containsSequence(args, "--bind", sshDir, sshDir) {
			t.Errorf("blocked dir %q must not appear as a parent-level mount; args: %v", sshDir, args)
		}
	})

	t.Run("unclassified dotdir is ro-bound", func(t *testing.T) {
		if !containsSequence(args, "--ro-bind", vimDir, vimDir) {
			t.Errorf("expected --ro-bind for unclassified dir %q; args: %v", vimDir, args)
		}
	})

	t.Run("non-dotfile is not mounted", func(t *testing.T) {
		readme := filepath.Join(home, "README.md")
		for _, a := range args {
			if a == readme {
				t.Errorf("non-dotfile %q must not appear in args; args: %v", readme, args)
			}
		}
	})

	t.Run("blocked sub-path gets tmpfs", func(t *testing.T) {
		if !containsSequence(args, "--tmpfs", secretDir) {
			t.Errorf("expected --tmpfs for blocked sub-path %q; args: %v", secretDir, args)
		}
	})

	t.Run("allowed sub-path gets rw-bind", func(t *testing.T) {
		if !containsSequence(args, "--bind", gooseDir, gooseDir) {
			t.Errorf("expected --bind for allowed sub-path %q; args: %v", gooseDir, args)
		}
	})

	t.Run("blocked sub-path tmpfs precedes allowed sub-path rw-bind", func(t *testing.T) {
		iBlocked := indexOfSequence(args, "--tmpfs", secretDir)
		iAllowed := indexOfSequence(args, "--bind", gooseDir, gooseDir)
		if iBlocked == -1 || iAllowed == -1 {
			t.Fatal("prerequisite sequences not found in args")
		}
		if iBlocked > iAllowed {
			t.Errorf("--tmpfs for blocked sub-path (idx %d) must come before --bind for allowed sub-path (idx %d)", iBlocked, iAllowed)
		}
	})
}

// containsSequence reports whether needle appears as a contiguous subsequence in haystack.
func containsSequence(haystack []string, needle ...string) bool {
	return indexOfSequence(haystack, needle...) != -1
}

// indexOfSequence returns the index of the first element of the first occurrence of
// needle as a contiguous subsequence in haystack, or -1 if not found.
func indexOfSequence(haystack []string, needle ...string) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j, v := range needle {
			if haystack[i+j] != v {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
