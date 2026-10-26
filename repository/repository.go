package repository

import (
	"errors"
	"os"
	"path/filepath"
)

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

// ─────────────────────────────────────────────────────────────────────────────
