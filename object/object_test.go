package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestHash(t *testing.T) {
	content := []byte("hello world")
	expected := fmt.Sprintf("blob %d\x00%s", len(content), string(content))
	expectedHash := sha1.Sum([]byte(expected))

	hash, payload := Hash([]byte(content))

	if string(payload) != expected {
		t.Errorf("payload does not match expected")
	}

	if !bytes.Equal(hash[:], expectedHash[:]) {
		t.Errorf("hash does not match expected")
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	content := []byte("hello world")
	payload := []byte(fmt.Sprintf("blob %d\x00%s", len(content), string(content)))
	hash := sha1.Sum(payload)
	hashString := hex.EncodeToString(hash[:])

	if err := Write(dir, hash, payload); err != nil {
		t.Fatalf("failed to write object: %v", err)
	}

	readContent, err := os.ReadFile(filepath.Join(dir, ".git", "objects", hashString[0:2], hashString[2:]))

	if err != nil {
		t.Fatalf("expected object file to exist: %v", err)
	}

	b := bytes.NewReader(readContent)
	testReader, err := zlib.NewReader(b)

	if err != nil {
		t.Fatalf("failed to create zlib reader: %v", err)
	}
	defer testReader.Close()

	fileContent, err := io.ReadAll(testReader)

	if err != nil {
		t.Errorf("failed to read object file bytes: %v", err)
	}

	if !bytes.Equal(fileContent, payload) {
		t.Errorf("file contents do not match expected")
	}
}
