package commit

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emerikaji/burrow/object"
)

type Signature struct {
	Name  string
	Email string
	When  time.Time
}
type Metadata struct {
	RootTreeHash       [20]byte
	ParentCommitHashes [][20]byte
	Author             Signature
	Committer          Signature
	Message            string
}

func (m *Metadata) PopulateParent(path string) (ref string, err error) {
	var content []byte
	if content, err = os.ReadFile(filepath.Join(path, ".git", "HEAD")); errors.Is(err, os.ErrNotExist) {
		err = nil
		return
	} else if err != nil {
		return
	}

	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return
	}

	if fields[0] == "ref:" && len(fields) < 2 {
		var hash []byte
		hash, err = hex.DecodeString(strings.TrimSpace(string(content)))
		if err != nil {
			return
		}

		var parentHash [20]byte
		copy(parentHash[:], hash)
		m.ParentCommitHashes = append(m.ParentCommitHashes, parentHash)

		return
	}

	ref = filepath.Join(path, ".git", fields[1])

	if content, err = os.ReadFile(ref); errors.Is(err, os.ErrNotExist) {
		err = nil
		return
	} else if err != nil {
		return
	}

	hash, err := hex.DecodeString(strings.TrimSpace(string(content)))
	if err != nil {
		return
	}

	var parentHash [20]byte
	copy(parentHash[:], hash)
	m.ParentCommitHashes = append(m.ParentCommitHashes, parentHash)

	return
}

func (m *Metadata) Write(path string) (commitHash [20]byte, err error) {
	var buf bytes.Buffer

	// Convert [20]byte to 40-char hex string
	buf.WriteString(fmt.Sprintf("tree %s\n", hex.EncodeToString(m.RootTreeHash[:])))

	for _, parent := range m.ParentCommitHashes {
		buf.WriteString(fmt.Sprintf("parent %s\n", hex.EncodeToString(parent[:])))
	}

	buf.WriteString(fmt.Sprintf("author %s <%s> %d %s\n", m.Author.Name, m.Author.Email, m.Author.When.Unix(), m.Author.When.Format("-0700")))
	buf.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n", m.Committer.Name, m.Committer.Email, m.Committer.When.Unix(), m.Committer.When.Format("-0700")))
	buf.WriteString(fmt.Sprintf("\n%s", m.Message))

	hash, payload := object.Hash(object.Commit, buf.Bytes())

	return hash, object.Write(path, hash, payload)
}

func ReadTree(path string, commitHash [20]byte) (treeHash [20]byte, err error) {
	content, err := object.ReadAndUnpack(path, commitHash)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "tree ") {
		err = fmt.Errorf("malformed commit object: missing tree header")
		return
	}

	treeHex := strings.TrimPrefix(lines[0], "tree ")
	decoded, err := hex.DecodeString(treeHex)
	if err != nil {
		return
	}

	copy(treeHash[:], decoded)
	return
}

// ─────────────────────────────────────────────────────────────────────────────
