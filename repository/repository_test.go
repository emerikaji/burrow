package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name         string
		dir          string
		headContent  string
		shouldReinit bool
	}{
		{
			name:         "empty directory",
			dir:          t.TempDir(),
			headContent:  "ref: refs/heads/main\n",
			shouldReinit: false,
		},
		{
			name:         "existing repository",
			dir:          t.TempDir(),
			headContent:  "this head exists",
			shouldReinit: true,
		},
	}

	gitExists := filepath.Join(tests[1].dir, ".git")
	headExists := filepath.Join(gitExists, "HEAD")

	if err := os.MkdirAll(gitExists, 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}
	if err := os.WriteFile(headExists, []byte(tests[1].headContent), 0644); err != nil {
		t.Fatalf("failed to seed HEAD: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if reinit, err := Init(tt.dir); err != nil {
				t.Fatalf("failed to initialize repository: %v", err)
			} else {
				if reinit != tt.shouldReinit {
					t.Errorf("reinit expectation mismatch: got %v, want %v", reinit, tt.shouldReinit)
				}
			}

			git := filepath.Join(tt.dir, ".git")

			dirs := []string{
				git,
				filepath.Join(git, "objects"),
				filepath.Join(git, "refs", "heads"),
				filepath.Join(git, "refs", "tags"),
			}

			for _, dir := range dirs {
				if info, err := os.Stat(dir); err != nil || !info.IsDir() {
					t.Errorf("expected directory %s to exist", dir)
				}
			}

			head := filepath.Join(git, "HEAD")

			content, err := os.ReadFile(head)

			if err != nil {
				t.Errorf("expected HEAD file to exist: %v", err)
			}

			if string(content) != tt.headContent {
				t.Errorf("expected HEAD content %q, got %q", tt.headContent, string(content))
			}
		})
	}
}
