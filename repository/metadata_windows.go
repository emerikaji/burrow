//go:build windows

package repository

import (
	"os"
	"syscall"
	"time"

	"github.com/emerikaji/burrow/index"
)

func populateSystemSpecifics(info os.FileInfo, entry *index.Entry) {
	if sys, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		entry.Created = time.Unix(0, sys.CreationTime.Nanoseconds())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
