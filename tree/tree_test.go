package tree

import (
	"testing"

	"github.com/emerikaji/burrow/index"
)

func TestBuild(t *testing.T) {
	idx := &index.Index{
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
			Path: "src/package/package.go",
		},
	}

	tree, err := Build(idx)
	if err != nil {
		t.Fatalf("failed to build tree: %s", err)
	}

	if _, ok := tree.Children["src"].Children["package"].Children["package.go"]; !ok {
		t.Errorf("package.go not found in tree")
	}
}
