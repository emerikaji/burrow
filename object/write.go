package object

import (
	"compress/zlib"
	"encoding/hex"
	"os"
	"path/filepath"
)

func Write(path string, hash [20]byte, payload []byte) (err error) {
	hashString := hex.EncodeToString(hash[:])

	objectPath := filepath.Join(path, ".git", "objects", hashString[0:2])
	object := filepath.Join(objectPath, hashString[2:])

	if _, err = os.Stat(object); err == nil {
		return
	}

	if err = os.MkdirAll(objectPath, 0755); err != nil {
		return
	}

	tmpObject, err := os.CreateTemp(objectPath, "tmp*")
	if err != nil {
		return
	}
	tmp := tmpObject.Name()

	defer func() {
		tmpObject.Close()
		if err != nil {
			os.Remove(tmp)
		}
	}()

	objectWriter := zlib.NewWriter(tmpObject)
	_, writerErr := objectWriter.Write(payload)
	closeErr := objectWriter.Close()

	err = tmpObject.Close()

	if writerErr != nil {
		return writerErr
	}

	if closeErr != nil {
		return closeErr
	}

	if err := os.Chmod(tmp, 0444); err != nil {
		return err
	}

	return os.Rename(tmp, object)
}

// ─────────────────────────────────────────────────────────────────────────────
