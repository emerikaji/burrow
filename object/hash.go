package object

import (
	"crypto/sha1"
	"fmt"
)

func Hash(content []byte) (hash [20]byte, payload []byte) {
	payload = append([]byte(fmt.Sprintf("blob %d\x00", len(content))), content...)

	hash = sha1.Sum(payload)

	return
}

// ─────────────────────────────────────────────────────────────────────────────
