package tree

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emerikaji/burrow/index"
)

var sampleIndex = &index.Index{
	index.Entry{
		Hash: [20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
		Path: "src/package/spackage/sspackage/sspackage.go",
	},
	index.Entry{
		Hash: [20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
		Path: "README.md",
	},
	index.Entry{
		Hash: [20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
		Path: "src/main.go",
	},
	index.Entry{
		Hash: [20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
		Path: "src/otherpackage/package.go",
	},
	index.Entry{
		Hash: [20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
		Path: "src/package/package.go",
	},
}

func TestBuild(t *testing.T) {
	leaves, err := Build(sampleIndex)
	if err != nil {
		t.Fatalf("failed to build tree: %s", err)
	}

	if len(leaves) != 2 {
		t.Errorf("got %d entries, expected 2", len(leaves))
	}

	if leaves[0].HasChildSubtrees() == true || leaves[0].HasChildSubtrees() == true {
		t.Errorf("expected only bottom subtrees")
	}
}

func TestWrite(t *testing.T) {
	leaves, err := Build(sampleIndex)
	if err != nil {
		t.Fatalf("failed to build tree: %s", err)
	}

	dir := t.TempDir()
	objects := filepath.Join(dir, ".git", "objects")
	if err = os.MkdirAll(objects, 0755); err != nil {
		t.Fatalf("failed to seed directory: %s", err)
	}

	rootHash, err := Write(dir, leaves)
	if err != nil {
		t.Fatalf("failed to write tree: %s", err)
	}

	rootHashHex := hex.EncodeToString(rootHash[:])

	// Recursive helper to inspect trees bottom-down/top-up via Git
	var verifyTree func(hashHex string, indent string)
	verifyTree = func(hashHex string, indent string) {
		cmd := exec.Command("git", "cat-file", "-p", hashHex)
		cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_OBJECT_DIRECTORY=%s", objects))

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git cat-file failed for hash %s: %v\nOutput: %s", hashHex, err, string(output))
		}

		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			line := scanner.Text()
			// Format: <mode> <type> <hash>\t<name>
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			mode, objType, objHash, name := parts[0], parts[1], parts[2], parts[3]
			t.Logf("%s[%s] %s (%s -> %s)", indent, mode, name, objType, objHash)

			// If it's a subdirectory tree, recursively verify it exists too!
			if objType == "tree" {
				verifyTree(objHash, indent+"  ")
			}
		}
	}

	t.Logf("Starting deep tree verification from root: %s", rootHashHex)
	verifyTree(rootHashHex, "")
}
