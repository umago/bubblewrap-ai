package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSplitFields(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "space separated string",
			in:   []string{"--ro-bind /var /var"},
			want: []string{"--ro-bind", "/var", "/var"},
		},
		{
			name: "already split array",
			in:   []string{"--ro-bind", "/var", "/var"},
			want: []string{"--ro-bind", "/var", "/var"},
		},
		{
			name: "mixed space separated and single tokens",
			in:   []string{"--unshare-net", "--ro-bind /var /var", "--unshare-ipc"},
			want: []string{"--unshare-net", "--ro-bind", "/var", "/var", "--unshare-ipc"},
		},
		{
			name: "empty input",
			in:   []string{},
			want: nil,
		},
		{
			name: "element with only spaces is dropped",
			in:   []string{"--unshare-net", "   ", "--unshare-ipc"},
			want: []string{"--unshare-net", "--unshare-ipc"},
		},
		{
			name: "multiple spaces between tokens",
			in:   []string{"--ro-bind    /var   /var"},
			want: []string{"--ro-bind", "/var", "/var"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitFields(tc.in)
			if !slices.Equal(got, tc.want) {
				t.Errorf("splitFields(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestLoadConfigSplitsFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	jsonContent := `{
  "bwrap_extra_args": [
    "--unshare-net",
    "--ro-bind /var /var"
  ]
}`
	if err := os.WriteFile(path, []byte(jsonContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	wantExtra := []string{"--unshare-net", "--ro-bind", "/var", "/var"}
	if !slices.Equal(cfg.BwrapExtraArgs, wantExtra) {
		t.Errorf("BwrapExtraArgs = %v, want %v", cfg.BwrapExtraArgs, wantExtra)
	}
}

func TestLoadConfigPreservesAlreadySplit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	jsonContent := `{
  "bwrap_extra_args": [
    "--ro-bind",
    "/var",
    "/var"
  ],
  "command": [
    "bash"
  ]
}`
	if err := os.WriteFile(path, []byte(jsonContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	wantExtra := []string{"--ro-bind", "/var", "/var"}
	if !slices.Equal(cfg.BwrapExtraArgs, wantExtra) {
		t.Errorf("BwrapExtraArgs = %v, want %v", cfg.BwrapExtraArgs, wantExtra)
	}
}
