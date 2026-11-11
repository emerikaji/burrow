package commit

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/emerikaji/burrow/object"
)

func TestPopulate(t *testing.T) {
	dir := t.TempDir() // Automatically isolated and cleaned up!
	git := filepath.Join(dir, ".git")
	heads := filepath.Join(git, "refs", "heads")
	main := filepath.Join(heads, "main")
	HEAD := filepath.Join(git, "HEAD")

	rootHash := [20]byte{0xab, 0xcd, 0xef, 0x12, 0x34}
	rootHashHex := hex.EncodeToString(rootHash[:])

	if err := os.MkdirAll(heads, 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	if err := os.WriteFile(main, []byte(rootHashHex+"\n"), 0644); err != nil {
		t.Fatalf("failed to seed main refs file: %v", err)
	}

	if err := os.WriteFile(HEAD, []byte("ref: refs/heads/main"), 0644); err != nil {
		t.Fatalf("failed to seed HEAD file: %v", err)
	}

	commitData := &Metadata{}
	if _, err := commitData.PopulateParent(dir); err != nil {
		t.Fatalf("failed to populate parents: %v", err)
	}

	if len(commitData.ParentCommitHashes) != 1 || commitData.ParentCommitHashes[0] != rootHash {
		t.Fatalf("parent hashes incorrectly populated: got %x", commitData.ParentCommitHashes)
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	objectsDir := filepath.Join(dir, ".git", "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	fixedTime := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	var dummyTree [20]byte
	copy(dummyTree[:], "abcdef12340000000000")

	commitData := &Metadata{
		RootTreeHash:       dummyTree,
		ParentCommitHashes: nil,
		Author: Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  fixedTime,
		},
		Committer: Signature{
			Name:  "Test Committer",
			Email: "committer@example.com",
			When:  fixedTime,
		},
		Message: "Initial test commit\n",
	}

	// 1. Write the commit using our package logic
	commitHash, err := commitData.Write(dir)
	if err != nil {
		t.Fatalf("failed to write commit: %v", err)
	}

	// 2. Read it back using our own object reader
	content, err := object.ReadAndUnpack(dir, commitHash)
	if err != nil {
		t.Fatalf("failed to read and unpack written commit object: %v", err)
	}

	// 3. Build the expected payload dynamically from the data fields
	var expectedBuf bytes.Buffer
	expectedBuf.WriteString(fmt.Sprintf("tree %s\n", hex.EncodeToString(commitData.RootTreeHash[:])))
	for _, parent := range commitData.ParentCommitHashes {
		expectedBuf.WriteString(fmt.Sprintf("parent %s\n", hex.EncodeToString(parent[:])))
	}
	expectedBuf.WriteString(fmt.Sprintf("author %s <%s> %d %s\n", commitData.Author.Name, commitData.Author.Email, commitData.Author.When.Unix(), commitData.Author.When.Format("-0700")))
	expectedBuf.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n", commitData.Committer.Name, commitData.Committer.Email, commitData.Committer.When.Unix(), commitData.Committer.When.Format("-0700")))
	expectedBuf.WriteString(fmt.Sprintf("\n%s", commitData.Message))

	expectedPayload := expectedBuf.String()

	// 4. Verify round-trip equality
	if string(content) != expectedPayload {
		t.Errorf("expected commit payload:\n%q\ngot:\n%q", expectedPayload, string(content))
	}
}

func TestReadTree(t *testing.T) {
	dir := t.TempDir()

	testTreeHash := [20]byte{0xab, 0xcd, 0xef, 0x12, 0x34}

	hash, payload := object.Hash(object.Commit, []byte("tree "+hex.EncodeToString(testTreeHash[:])))
	if err := object.Write(dir, hash, payload); err != nil {
		t.Fatalf("failed to write object: %v", err)
	}

	treeHash, err := ReadTree(dir, hash)
	if err != nil {
		t.Fatalf("failed to read commit: %v", err)
	}

	if treeHash != testTreeHash {
		t.Fatalf("tree hash mismatch: got %x, want %x", treeHash, testTreeHash)
	}
}
