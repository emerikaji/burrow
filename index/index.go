package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var (
	signature    = [4]byte{'D', 'I', 'R', 'C'}
	errorSmall   = errors.New("index file smaller than expected")
	errorCorrupt = errors.New("index file corrupt")
)

type Entry struct {
	Created     time.Time
	Modified    time.Time
	Dev         uint32
	Ino         uint32
	Mode        os.FileMode
	UID         uint32
	GID         uint32
	Size        uint32
	Hash        [20]byte
	Stage       uint8
	AssumeValid bool
	Path        string
}

type entryBytes struct {
	CreatedSeconds      uint32
	CreatedNanoseconds  uint32
	ModifiedSeconds     uint32
	ModifiedNanoseconds uint32
	Dev                 uint32
	Ino                 uint32
	Mode                os.FileMode
	UID                 uint32
	GID                 uint32
	Size                uint32
	Hash                [20]byte
	Flags               uint16
}

type Index []Entry

type indexBytes struct {
	Signature [4]byte
	Version   uint32
	Entries   uint32
}

func Read(path string) (index *Index, err error) {
	index = &Index{}

	var content []byte
	if content, err = os.ReadFile(filepath.Join(path, ".git", "index")); errors.Is(err, os.ErrNotExist) {
		err = nil
		return
	} else if err != nil {
		return
	}

	if len(content) < 32 {
		err = errorSmall
		return
	}

	var indexBytes indexBytes
	if _, err = binary.Decode(content[0:12], binary.BigEndian, &indexBytes); err != nil {
		return
	}

	if indexBytes.Signature != signature {
		log.Printf("bad signature 0x%d", binary.BigEndian.Uint32(indexBytes.Signature[:]))
		err = errorCorrupt
		return
	}

	if indexBytes.Version != 2 {
		log.Printf("bad index version %d", indexBytes.Version)
		err = errorCorrupt
		return
	}

	fileChecksum := content[len(content)-20:]
	calculatedChecksum := sha1.Sum(content[:len(content)-20])
	if !bytes.Equal(fileChecksum, calculatedChecksum[:]) {
		log.Printf("checksum mismatch")
		err = errorCorrupt
		return
	}

	offset := 12
	for i := 0; i < int(indexBytes.Entries); i++ {
		reader := bytes.NewReader(content[offset:])

		var entryBytes entryBytes
		if err = binary.Read(reader, binary.BigEndian, &entryBytes); err != nil {
			return
		}

		pathLen := int(entryBytes.Flags & 0x0FFF)

		pathBytes := make([]byte, pathLen)
		if err = binary.Read(reader, binary.BigEndian, &pathBytes); err != nil {
			return
		}

		index.Add(Entry{
			time.Unix(int64(entryBytes.CreatedSeconds), int64(entryBytes.CreatedNanoseconds)),
			time.Unix(int64(entryBytes.ModifiedSeconds), int64(entryBytes.ModifiedNanoseconds)),
			entryBytes.Dev,
			entryBytes.Ino,
			entryBytes.Mode,
			entryBytes.UID,
			entryBytes.GID,
			entryBytes.Size,
			entryBytes.Hash,
			uint8((entryBytes.Flags >> 12) & 0x03),
			(entryBytes.Flags & 0x8000) != 0,
			string(pathBytes),
		})

		offset += (62 + pathLen + 1 + 7) &^ 7
	}

	return
}

func (index *Index) Add(entry Entry) {
	i, found := slices.BinarySearchFunc(*index, entry, func(a, b Entry) int {
		return strings.Compare(a.Path, b.Path)
	})

	if found {
		(*index)[i] = entry
	} else {
		*index = slices.Insert(*index, i, entry)
	}
}

func (index *Index) Write(path string) (err error) {
	var buf bytes.Buffer

	if err = binary.Write(&buf, binary.BigEndian, indexBytes{
		signature,
		2,
		uint32(len(*index)),
	}); err != nil {
		return
	}

	for _, entry := range *index {
		pathLen := min(len(entry.Path), 0xFFF)
		var flags = uint16(pathLen & 0x0FFF)
		flags |= uint16(entry.Stage&0x03) << 12
		if entry.AssumeValid {
			flags |= 0x8000
		}

		if err = binary.Write(&buf, binary.BigEndian, entryBytes{
			uint32(entry.Created.Unix()),
			uint32(entry.Created.Nanosecond()),
			uint32(entry.Modified.Unix()),
			uint32(entry.Modified.Nanosecond()),
			entry.Dev,
			entry.Ino,
			entry.Mode,
			entry.UID,
			entry.GID,
			entry.Size,
			entry.Hash,
			flags,
		}); err != nil {
			return
		}

		if err = binary.Write(&buf, binary.BigEndian, []byte(entry.Path+"\x00")); err != nil {
			return
		}

		paddingNeeded := (8 - ((62 + pathLen + 1) % 8)) % 8
		if paddingNeeded > 0 {
			if err = binary.Write(&buf, binary.BigEndian, make([]byte, paddingNeeded)); err != nil {
				return
			}
		}
	}

	checksum := sha1.Sum(buf.Bytes())
	buf.Write(checksum[:])

	git := filepath.Join(path, ".git")

	tmp, err := os.CreateTemp(git, "index_tmp*")
	if err != nil {
		return
	}

	_, writerErr := tmp.Write(buf.Bytes())

	err = tmp.Close()

	if err != nil {
		os.Remove(tmp.Name())

		if writerErr == nil {
			err = writerErr
		}

		return
	}

	return os.Rename(tmp.Name(), filepath.Join(git, "index"))
}

// ─────────────────────────────────────────────────────────────────────────────
