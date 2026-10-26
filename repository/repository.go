package repository

import (
	"errors"
	"os"
	"path/filepath"
)

var NoRepoError = errors.New("not a git repository (or any of the parent directories)")

func CheckForRepoPwd() bool {
	pwd, err := os.Getwd()
	if err != nil {
		return false
	}

	return CheckForRepo(pwd)
}

func CheckForRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")

	if _, err := os.Stat(gitPath); err == nil {
		_, errHEAD := os.ReadFile(filepath.Join(gitPath, "HEAD"))
		info, errObjects := os.Stat(filepath.Join(gitPath, "objects"))
		if errHEAD == nil && errObjects == nil && info.IsDir() {
			return true
		}
	}

	nextPath := filepath.Dir(path)

	if path == nextPath {
		return false
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

func Add(paths []string) error {
	if !CheckForRepoPwd() {
		return NoRepoError
	}
	//TODO
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
