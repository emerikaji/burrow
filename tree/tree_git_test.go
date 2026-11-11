package tree

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"os/exec"
	"strings"
	"testing"
)

func TestGitWrite(t *testing.T) {
	dir := t.TempDir()

	initCmd := exec.Command("git", "init")
	initCmd.Dir = dir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	leaves, err := Build(sampleIndex)
	if err != nil {
		t.Fatalf("failed to build tree: %s", err)
	}

	rootHash, err := Write(dir, leaves)
	if err != nil {
		t.Fatalf("failed to write tree: %s", err)
	}

	rootHashHex := hex.EncodeToString(rootHash[:])

	// Ask real Git to recursively list all files in the tree (-r) with modes and hashes
	cmd := exec.Command("git", "ls-tree", "-r", rootHashHex)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-tree failed: %v\nOutput: %s", err, string(output))
	}

	// Parse git's output lines: <mode> <type> <hash>\t<path>
	gitEntries := make(map[string]string) // path -> mode
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		meta := strings.Fields(parts[0])
		if len(meta) < 3 {
			continue
		}
		mode, path := meta[0], parts[1]
		gitEntries[path] = mode
	}

	// Verify count and presence of every index entry
	if len(gitEntries) != len(*sampleIndex) {
		t.Errorf("git ls-tree entry count mismatch: expected %d, got %d", len(*sampleIndex), len(gitEntries))
	}

	for _, entry := range *sampleIndex {
		gitMode, exists := gitEntries[entry.Path]
		if !exists {
			t.Errorf("path present in index but missing from git tree: %s", entry.Path)
			continue
		}

		if gitMode != "100644" {
			t.Errorf("mode mismatch for %s: expected %s, got %s", entry.Path, "100644", gitMode)
		}
	}
}
