package repository

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/emerikaji/burrow/index"
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

func TestAdd(t *testing.T) {
	dir := t.TempDir()
	readmeFile := filepath.Join(dir, "README.md")

	git := filepath.Join(dir, ".git")
	objects := filepath.Join(git, "objects")
	headFile := filepath.Join(git, "HEAD")

	// set up repo
	if err := os.MkdirAll(objects, 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("failed to seed HEAD: %v", err)
	}

	// set up project files
	if err := os.WriteFile(readmeFile, []byte("# This is a test README!\nThere are some contents in it."), 0644); err != nil {
		t.Fatalf("failed to seed README.md: %v", err)
	}

	// burrow add
	if err := Add(dir, []string{"README.md"}); err != nil {
		t.Fatalf("failed to add README.md: %v", err)
	}

	// check results
	var err error
	var readIndex *index.Index
	if readIndex, err = index.Read(dir); err != nil {
		t.Fatalf("failed to read index: %v", err)
	}
	hashString := hex.EncodeToString((*readIndex)[0].Hash[:])
	if _, err = os.ReadFile(filepath.Join(objects, hashString[:2], hashString[2:])); err != nil {
		t.Fatalf("failed to read object: %v", err)
	}
}
