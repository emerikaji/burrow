package tree

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emerikaji/burrow/index"
	"github.com/emerikaji/burrow/object"
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

	// Helper to recursively parse binary tree objects and collect all full file paths
	var collectPaths func(hash [20]byte, prefix string) map[string]bool
	collectPaths = func(hash [20]byte, prefix string) map[string]bool {
		paths := make(map[string]bool)
		content, err := object.ReadAndUnpack(dir, hash)
		if err != nil {
			t.Fatalf("failed to unpack tree object: %v", err)
		}

		// Git tree format: sequence of entries: <mode> <name>\x00<20-byte binary hash>
		i := 0
		for i < len(content) {
			spaceIdx := bytes.IndexByte(content[i:], ' ')
			if spaceIdx == -1 {
				break
			}
			spaceIdx += i

			nullIdx := bytes.IndexByte(content[spaceIdx:], '\x00')
			if nullIdx == -1 {
				break
			}
			nullIdx += spaceIdx

			modeStr := string(content[i:spaceIdx])
			name := string(content[spaceIdx+1 : nullIdx])

			var entryHash [20]byte
			copy(entryHash[:], content[nullIdx+1:nullIdx+1+20])

			fullPath := name
			if prefix != "" {
				fullPath = prefix + "/" + name
			}

			// Git tree mode for subdirectories starts with "40" (e.g. "40000" or "040000")
			if strings.HasPrefix(modeStr, "40") || strings.HasPrefix(modeStr, "04") {
				subPaths := collectPaths(entryHash, fullPath)
				for p := range subPaths {
					paths[p] = true
				}
			} else {
				paths[fullPath] = true
			}

			i = nullIdx + 1 + 20
		}
		return paths
	}

	// Unpack and collect all paths written by our tree engine
	writtenPaths := collectPaths(rootHash, "")

	// Verify that every single entry in sampleIndex is present and accounted for
	if len(writtenPaths) != len(*sampleIndex) {
		t.Errorf("path count mismatch: expected %d entries from index, got %d in written tree", len(*sampleIndex), len(writtenPaths))
	}

	for _, entry := range *sampleIndex {
		if !writtenPaths[entry.Path] {
			t.Errorf("index entry path missing from written tree: %s", entry.Path)
		}
	}
}
