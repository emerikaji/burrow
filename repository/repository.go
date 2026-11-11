package repository

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emerikaji/burrow/commit"
	"github.com/emerikaji/burrow/index"
	"github.com/emerikaji/burrow/object"
	"github.com/emerikaji/burrow/tree"
	"golang.org/x/sync/errgroup"
)

var (
	errorNoRepo = errors.New("not a git repository (or any of the parent directories)")
)

func CheckForRepo(path string) (found bool, foundPath string) {
	gitPath := filepath.Join(path, ".git")

	if _, err := os.Stat(gitPath); err == nil {
		_, errHEAD := os.ReadFile(filepath.Join(gitPath, "HEAD"))
		info, errHeads := os.Stat(filepath.Join(gitPath, "refs", "heads"))
		if errHEAD == nil && errHeads == nil && info.IsDir() {
			return true, path
		}
	}

	nextPath := filepath.Dir(path)

	if path == nextPath {
		return false, ""
	}

	return CheckForRepo(nextPath)
}

func Init(path string) (reinit bool, err error) {
	git := filepath.Join(path, ".git")
	head := filepath.Join(git, "HEAD")
	refs := filepath.Join(git, "refs")

	if err = os.MkdirAll(filepath.Join(git, "objects"), 0755); err != nil {
		return
	}

	if err = os.MkdirAll(filepath.Join(refs, "heads"), 0755); err != nil {
		return
	}

	if err = os.MkdirAll(filepath.Join(refs, "tags"), 0755); err != nil {
		return
	}

	var f *os.File
	if f, err = os.OpenFile(head, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644); err != nil {
		if errors.Is(err, os.ErrExist) {
			return true, nil
		}
		return
	}

	_, writeErr := f.WriteString("ref: refs/heads/main\n")

	err = f.Close()

	if writeErr != nil {
		err = writeErr
	}

	return
}

type shared struct {
	group *errgroup.Group
	mu    sync.Mutex
	index *index.Index
	path  string
}

func Add(path string, filePaths []string) (err error) {
	s := &shared{}

	var found bool
	found, s.path = CheckForRepo(path)
	if !found {
		return errorNoRepo
	}

	s.group, _ = errgroup.WithContext(context.Background())

	if s.index, err = index.Read(path); err != nil {
		return
	}

	for _, filePath := range filePaths {
		filePath = filepath.Join(path, filePath)
		s.group.Go(func() error { return GatherEntries(s, filePath) })
	}

	if err = s.group.Wait(); err != nil {
		return
	}

	return s.index.Write(s.path)
}

func translateMode(mode os.FileMode) os.FileMode {
	if mode&os.ModeSymlink != 0 {
		return 0120000
	}

	if mode&0100 != 0 {
		return 0100755
	}

	return 0100644
}

func GatherEntries(s *shared, filePath string) (err error) {
	var info os.FileInfo
	if info, err = os.Lstat(filePath); err != nil {
		return
	}

	if info.Name() == ".git" {
		return
	}

	if info.IsDir() {
		var dirEntries []os.DirEntry
		if dirEntries, err = os.ReadDir(filePath); err != nil {
			return
		}

		for _, dirEntry := range dirEntries {
			s.group.Go(func() error { return GatherEntries(s, filepath.Join(filePath, dirEntry.Name())) })
		}

		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	hash, payload := object.Hash(object.Blob, content)

	if err = object.Write(s.path, hash, payload); err != nil {
		return
	}

	relativePath, err := filepath.Rel(s.path, filePath)
	if err != nil {
		return
	}

	entry := index.Entry{
		Modified:    info.ModTime(),
		Mode:        translateMode(info.Mode()),
		Size:        uint32(info.Size()),
		Hash:        hash,
		Stage:       0,
		AssumeValid: false,
		Path:        filepath.ToSlash(relativePath),
	}

	populateSystemSpecifics(info, &entry)

	s.mu.Lock()
	s.index.Add(entry)
	s.mu.Unlock()

	return
}

func Commit(path string, author commit.Signature, committer commit.Signature, message string) (commitHashHex string, branch string, isRoot bool, err error) {
	found, basePath := CheckForRepo(path)
	if !found {
		err = errorNoRepo
		return
	}

	idx, err := index.Read(basePath)
	if err != nil {
		return
	}

	leaves, err := tree.Build(idx)
	if err != nil {
		return
	}

	rootTreeHash, err := tree.Write(basePath, leaves)
	if err != nil {
		return
	}

	commitData := commit.Metadata{
		RootTreeHash: rootTreeHash,
		Author:       author,
		Committer:    committer,
		Message:      message,
	}

	ref, err := commitData.PopulateParent(path)
	if err != nil {
		return
	}

	log.Println(ref)

	_, branch, _ = strings.CutLast(ref, string(filepath.Separator))
	isRoot = len(commitData.ParentCommitHashes) == 0

	if len(commitData.ParentCommitHashes) > 0 {
		parentHash := commitData.ParentCommitHashes[0]

		var parentTreeHash [20]byte
		if parentTreeHash, err = commit.ReadTree(basePath, parentHash); err != nil {
			return
		}

		// If the new tree matches the parent tree, nothing changed!
		if rootTreeHash == parentTreeHash {
			err = fmt.Errorf("no changes added to commit (use 'burrow add' first)")
			return
		}
	}

	commitHash, err := commitData.Write(path)
	if err != nil {
		return
	}

	commitHashHex = hex.EncodeToString(commitHash[:])

	err = os.WriteFile(ref, []byte(commitHashHex+"\n"), 0644)

	return
}

// ─────────────────────────────────────────────────────────────────────────────
