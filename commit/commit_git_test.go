package commit

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestGitWrite(t *testing.T) {
	dir := t.TempDir()
	objects := filepath.Join(dir, ".git", "objects")

	if err := os.MkdirAll(objects, 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	fixedTime := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	var dummyTree [20]byte
	copy(dummyTree[:], "abcdef12340000000000")

	commitData := &Metadata{
		RootTreeHash:       dummyTree,
		ParentCommitHashes: nil,
		Author: Signature{
			Name:  "Git Test Author",
			Email: "gitauthor@example.com",
			When:  fixedTime,
		},
		Committer: Signature{
			Name:  "Git Test Committer",
			Email: "gitcommitter@example.com",
			When:  fixedTime,
		},
		Message: "Integration test commit message\n",
	}

	commitHash, err := commitData.Write(dir)
	if err != nil {
		t.Fatalf("failed to write commit: %v", err)
	}

	hashHex := hex.EncodeToString(commitHash[:])

	// Execute real git to inspect our generated commit object raw content
	cmd := exec.Command("git", "cat-file", "commit", hashHex)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_OBJECT_DIRECTORY=%s", objects))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git cat-file failed: %v\nOutput: %s", err, string(output))
	}

	// Build expected payload dynamically to match what we fed the method
	var expectedBuf bytes.Buffer
	expectedBuf.WriteString(fmt.Sprintf("tree %s\n", hex.EncodeToString(commitData.RootTreeHash[:])))
	for _, parent := range commitData.ParentCommitHashes {
		expectedBuf.WriteString(fmt.Sprintf("parent %s\n", hex.EncodeToString(parent[:])))
	}
	expectedBuf.WriteString(fmt.Sprintf("author %s <%s> %d %s\n", commitData.Author.Name, commitData.Author.Email, commitData.Author.When.Unix(), commitData.Author.When.Format("-0700")))
	expectedBuf.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n", commitData.Committer.Name, commitData.Committer.Email, commitData.Committer.When.Unix(), commitData.Committer.When.Format("-0700")))
	expectedBuf.WriteString(fmt.Sprintf("\n%s", commitData.Message))

	expectedPayload := expectedBuf.String()

	// Verify exact byte-for-byte equality with real Git's parsing view
	if string(output) != expectedPayload {
		t.Errorf("git output mismatch!\nExpected:\n%q\nGot:\n%q", expectedPayload, string(output))
	}
}
