package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func createDummyIndex(t *testing.T, entry *Entry) []byte {
	t.Helper()

	pathLen := min(len(entry.Path), 0xFFF)
	var flags = uint16(pathLen & 0x0FFF)
	flags |= uint16(entry.Stage&0x03) << 12
	if entry.AssumeValid {
		flags |= 0x8000
	}

	var buf bytes.Buffer
	// index
	buf.Write([]byte("DIRC"))                       // header
	binary.Write(&buf, binary.BigEndian, uint32(2)) // version 2
	binary.Write(&buf, binary.BigEndian, uint32(1)) // entries
	// entry
	entryStart := buf.Len()
	binary.Write(&buf, binary.BigEndian, uint32(entry.Created.Unix()))        // ctime seconds
	binary.Write(&buf, binary.BigEndian, uint32(entry.Created.Nanosecond()))  // ctime nanoseconds
	binary.Write(&buf, binary.BigEndian, uint32(entry.Modified.Unix()))       // mtime seconds
	binary.Write(&buf, binary.BigEndian, uint32(entry.Modified.Nanosecond())) // mtime nanoseconds
	binary.Write(&buf, binary.BigEndian, entry.Dev)                           // dev
	binary.Write(&buf, binary.BigEndian, entry.Ino)                           // ino
	binary.Write(&buf, binary.BigEndian, entry.Mode)                          // mode (regular file)
	binary.Write(&buf, binary.BigEndian, entry.UID)                           // uid
	binary.Write(&buf, binary.BigEndian, entry.GID)                           // gid
	binary.Write(&buf, binary.BigEndian, entry.Size)                          // file size (bytes)
	buf.Write(entry.Hash[:])                                                  //hash
	binary.Write(&buf, binary.BigEndian, flags)                               //flags
	buf.Write([]byte(entry.Path + "\x00"))                                    // filename
	// padding
	entryLen := buf.Len() - entryStart
	paddingNeeded := 8 - (entryLen % 8)
	if paddingNeeded < 8 {
		buf.Write(make([]byte, paddingNeeded))
	}
	// checksum
	checksum := sha1.Sum(buf.Bytes())
	buf.Write(checksum[:])

	return buf.Bytes()
}

func TestRead(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	if index, err := Read(dir); err != nil {
		t.Fatalf("read on empty index failed: %v", err)
	} else if index == nil || len(*index) > 0 {
		t.Errorf("invalid index from empty: %v", *index)
	}

	testEntry := Entry{
		time.Unix(1600000000, 0),
		time.Unix(1600000000, 0),
		12,
		34,
		0100644,
		1000,
		1000,
		13,
		[20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
		1,
		true,
		"foo.txt",
	}

	if err := os.WriteFile(filepath.Join(dir, ".git", "index"), createDummyIndex(t, &testEntry), 0644); err != nil {
		t.Fatalf("failed to seed index file: %v", err)
	}

	if index, err := Read(dir); err != nil {
		t.Fatalf("read on full index failed: %v,%v", err, *index)
	} else if index == nil || len(*index) == 0 || (*index)[0] != testEntry {
		t.Errorf("invalid index from existing: %v", *index)
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	index := &Index{
		Entry{
			time.Unix(1600000000, 0),
			time.Unix(1600000000, 0),
			12,
			34,
			0100644,
			1000,
			1000,
			13,
			[20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
			0,
			false,
			"bar.txt",
		},
		Entry{
			time.Unix(1600000000, 0),
			time.Unix(1600000000, 0),
			12,
			34,
			0100644,
			1000,
			1000,
			13,
			[20]byte{0xab, 0xcd, 0xef, 0x12, 0x34},
			0,
			false,
			"foo.txt",
		},
	}

	if err := index.Write(dir); err != nil {
		t.Fatalf("failed to write index: %v", err)
	}

	readIndex, err := Read(dir)
	if err != nil {
		t.Fatalf("read on index failed: %v", err)
	}

	if !reflect.DeepEqual(*index, *readIndex) {
		t.Fatalf("read index does not match expected index: %v,%v", *readIndex, *index)
	}
}
