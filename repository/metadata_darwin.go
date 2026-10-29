//go:build darwin

package repository

import (
	"os"
	"time"

	"github.com/emerikaji/burrow/index"
)

func populateSystemSpecifics(info os.FileInfo, entry *index.Entry) {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		entry.Dev = uint32(sys.Dev)
		entry.Ino = uint32(sys.Ino)
		entry.UID = sys.Uid
		entry.GID = sys.Gid
		entry.Created = time.Unix(int64(sys.Ctimespec.Sec), int64(sys.Ctimespec.Nsec))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
