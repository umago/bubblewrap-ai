package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

const releasesAPI = "https://api.github.com/repos/umago/bubblewrap-ai/releases/latest"

const maxBinarySize = 10 << 20 // 10 MB

var (
	apiClient      = &http.Client{Timeout: 30 * time.Second}
	downloadClient = &http.Client{Timeout: 120 * time.Second}
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

// runUpdate downloads the latest bwai binary from GitHub releases,
// verifies its SHA-256 digest, and replaces the current binary in-place.
func runUpdate() {
	currentExe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bwai: cannot determine current executable: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Checking for latest release...")

	release, err := fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bwai: %v\n", err)
		os.Exit(1)
	}

	asset := findAsset(release.Assets, "bwai")
	if asset == nil {
		fmt.Fprintf(os.Stderr, "bwai: no 'bwai' asset found in release %s\n", release.TagName)
		os.Exit(1)
	}

	fmt.Printf("Downloading bwai %s...\n", release.TagName)

	data, err := downloadAsset(asset.BrowserDownloadURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bwai: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded %d bytes.\n", len(data))

	if asset.Digest == "" {
		fmt.Fprintln(os.Stderr, "bwai: no digest available — refusing to install unverified binary")
		os.Exit(1)
	}
	if err := verifyDigest(data, asset.Digest); err != nil {
		fmt.Fprintf(os.Stderr, "bwai: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("SHA-256 digest verified.")

	if err := replaceBinary(data, currentExe); err != nil {
		fmt.Fprintf(os.Stderr, "bwai: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("bwai updated successfully (%s → %s).\n", version, release.TagName)
}

func fetchLatestRelease() (*githubRelease, error) {
	resp, err := apiClient.Get(releasesAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release response: %w", err)
	}
	return &release, nil
}

func findAsset(assets []githubAsset, name string) *githubAsset {
	for i := range assets {
		if assets[i].Name == name {
			return &assets[i]
		}
	}
	return nil
}

func downloadAsset(url string) ([]byte, error) {
	resp, err := downloadClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBinarySize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if int64(len(data)) > maxBinarySize {
		return nil, fmt.Errorf("downloaded binary exceeds max size of %d bytes", maxBinarySize)
	}
	return data, nil
}

// verifyDigest compares the SHA-256 of data against the expected digest.
// The digest from GitHub is in the format "sha256:hex...".
func verifyDigest(data []byte, expected string) error {
	colon := strings.Index(expected, ":")
	if colon >= 0 {
		expected = expected[colon+1:]
	}

	hash := sha256.Sum256(data)
	actual := hex.EncodeToString(hash[:])

	if actual != expected {
		return fmt.Errorf("digest mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

// replaceBinary writes the new binary data to the current executable path.
// The current binary is renamed to .old first to avoid ETXTBSY on Linux.
func replaceBinary(data []byte, currentExe string) error {
	oldExe := currentExe + ".old"

	// Clean up leftover from a previous failed update
	if _, err := os.Stat(oldExe); err == nil {
		if err := os.Remove(oldExe); err != nil {
			return fmt.Errorf("failed to remove stale backup %s: %w", oldExe, err)
		}
	}

	// Rename current binary out of the way
	if _, err := os.Stat(currentExe); err == nil {
		if err := os.Rename(currentExe, oldExe); err != nil {
			return fmt.Errorf("failed to rename %s: %w", currentExe, err)
		}
	}

	// Write the new binary
	if err := os.WriteFile(currentExe, data, 0o755); err != nil {
		// Try to restore the old binary
		if _, statErr := os.Stat(oldExe); statErr == nil {
			if renameErr := os.Rename(oldExe, currentExe); renameErr != nil {
				fmt.Fprintf(os.Stderr, "bwai: CRITICAL — failed to restore backup after write error: %v\n", renameErr)
			}
		}
		return fmt.Errorf("failed to write new binary to %s: %w", currentExe, err)
	}

	// Preserve ownership from the old binary
	if info, err := os.Stat(oldExe); err == nil {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			if err := os.Chown(currentExe, int(stat.Uid), int(stat.Gid)); err != nil {
				if renameErr := os.Rename(oldExe, currentExe); renameErr != nil {
					fmt.Fprintf(os.Stderr, "bwai: CRITICAL — failed to restore backup after chown error: %v\n", renameErr)
				}
				return fmt.Errorf("failed to chown new binary: %w", err)
			}
		}
	}

	// Delete the old backup
	_ = os.Remove(oldExe)

	return nil
}
