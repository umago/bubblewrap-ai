package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyDigest_MatchingSHA256(t *testing.T) {
	data := []byte("hello world")
	hash := sha256.Sum256(data)
	expected := "sha256:" + hex.EncodeToString(hash[:])

	if err := verifyDigest(data, expected); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestVerifyDigest_BareHex(t *testing.T) {
	data := []byte("hello world")
	hash := sha256.Sum256(data)
	expected := hex.EncodeToString(hash[:])

	if err := verifyDigest(data, expected); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestVerifyDigest_Mismatch(t *testing.T) {
	data := []byte("hello world")
	expected := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	err := verifyDigest(data, expected)
	if err == nil {
		t.Fatal("expected error on digest mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "digest mismatch") {
		t.Errorf("error should mention 'digest mismatch', got: %v", err)
	}
}

func TestFindAsset_Found(t *testing.T) {
	assets := []githubAsset{
		{Name: "checksums.txt"},
		{Name: "bwai", BrowserDownloadURL: "https://example.com/bwai"},
	}
	a := findAsset(assets, "bwai")
	if a == nil {
		t.Fatal("expected to find 'bwai' asset")
	}
	if a.BrowserDownloadURL != "https://example.com/bwai" {
		t.Errorf("unexpected URL: %s", a.BrowserDownloadURL)
	}
}

func TestFindAsset_NotFound(t *testing.T) {
	assets := []githubAsset{
		{Name: "checksums.txt"},
	}
	if findAsset(assets, "bwai") != nil {
		t.Fatal("expected nil for missing asset")
	}
}

func TestReplaceBinary_Basic(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "bwai")

	if err := os.WriteFile(current, []byte("old version"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceBinary([]byte("new version"), current); err != nil {
		t.Fatalf("replaceBinary failed: %v", err)
	}

	content, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new version" {
		t.Errorf("expected 'new version', got %q", content)
	}

	// Old backup should be cleaned up
	if _, err := os.Stat(current + ".old"); !os.IsNotExist(err) {
		t.Error("old backup should have been removed")
	}
}

func TestReplaceBinary_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "bwai")

	if err := replaceBinary([]byte("fresh install"), current); err != nil {
		t.Fatalf("replaceBinary failed: %v", err)
	}

	content, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "fresh install" {
		t.Errorf("expected 'fresh install', got %q", content)
	}
}

func TestReplaceBinary_Permissions(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "bwai")

	if err := os.WriteFile(current, []byte("old version"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := replaceBinary([]byte("new version"), current); err != nil {
		t.Fatalf("replaceBinary failed: %v", err)
	}

	info, err := os.Stat(current)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("expected 0o755 permissions, got %o", info.Mode().Perm())
	}
}

func TestReplaceBinary_CleansStaleBackup(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "bwai")

	if err := os.WriteFile(current, []byte("current"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(current+".old", []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := replaceBinary([]byte("new"), current); err != nil {
		t.Fatalf("replaceBinary failed: %v", err)
	}

	if _, err := os.Stat(current + ".old"); !os.IsNotExist(err) {
		t.Error("stale backup should have been removed before update")
	}
}
