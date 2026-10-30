package tree

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/emerikaji/burrow/index"
	"github.com/emerikaji/burrow/object"
)

type Subtree interface {
	Parent() Subtree
	HasChildSubtrees() bool
	Goto(name string) Subtree
	AddEntry(entry Entry)
	Payload() []byte
	SetHash(hash [20]byte)
}

type Entry interface {
	Name() string
	Hash() [20]byte
}

type Directory struct {
	name                string
	hash                [20]byte
	parent              Subtree
	entries             map[string]Entry
	hasDirectoryEntries bool
}

func (d *Directory) Parent() Subtree {
	return d.parent
}

func (d *Directory) HasChildSubtrees() bool {
	return d.hasDirectoryEntries
}

func (d *Directory) Goto(name string) Subtree {
	directory, ok := d.entries[name+"/"]
	if !ok {
		return nil
	}

	switch directory.(type) {
	case *Directory:
		return directory.(*Directory)
	}

	return nil
}

func (d *Directory) AddEntry(entry Entry) {
	key := entry.Name()
	switch entry.(type) {
	case *Directory:
		d.hasDirectoryEntries = true
		key += "/"
	}

	d.entries[key] = entry
}

func (d *Directory) Payload() []byte {
	var buf bytes.Buffer

	keys := slices.Sorted(maps.Keys(d.entries))

	entries := make([]Entry, len(keys))
	for i, key := range keys {
		entries[i] = d.entries[key]
	}

	var mode string
	for _, e := range entries {
		switch e.(type) {
		case *Directory:
			mode = "40000"
		case *File:
			mode = "100644"
		default:
			panic("this REALLY should not happen")
		}
		hash := e.Hash()

		buf.Write([]byte(fmt.Sprintf("%s %s\x00", mode, e.Name())))
		buf.Write(hash[:])
	}

	return buf.Bytes()
}

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) SetHash(hash [20]byte) {
	d.hash = hash
}

func (d *Directory) Hash() [20]byte {
	return d.hash
}

type File struct {
	name string
	hash [20]byte
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Hash() [20]byte {
	return f.hash
}

func Build(idx *index.Index) (leaves []Subtree, err error) {
	root := &Directory{
		name:    ".",
		entries: make(map[string]Entry),
	}
	leaves = []Subtree{root}
	leafIndex := 0

	var currentSubtree Subtree
	for _, indexEntry := range *idx {
		currentSubtree = root

		pathSplit := strings.Split(indexEntry.Path, "/")

		for i, name := range pathSplit {
			if i == len(pathSplit)-1 {
				currentSubtree.AddEntry(&File{
					name: name,
					hash: indexEntry.Hash,
				})

				continue
			}

			nextSubtree := currentSubtree.Goto(name)

			if nextSubtree == nil {
				newDir := &Directory{
					name:    name,
					parent:  currentSubtree,
					entries: make(map[string]Entry),
				}

				nextSubtree = newDir

				if currentSubtree.HasChildSubtrees() {
					leafIndex++
					leaves = append(leaves, nil)
				}
				leaves[leafIndex] = nextSubtree

				currentSubtree.AddEntry(newDir)
			}

			currentSubtree = nextSubtree

		}
	}

	return
}

func Write(path string, leaves []Subtree) (rootHash [20]byte, err error) {
	for len(leaves) > 0 {
		newLeaves := make([]Subtree, 0, len(leaves))
		for _, s := range leaves {
			hash, payload := object.Hash(object.Tree, s.Payload())
			s.SetHash(hash)

			if err = object.Write(path, hash, payload); err != nil {
				return
			}

			if s.Parent() == nil {
				rootHash = hash
				continue
			}

			if slices.Contains(newLeaves, s.Parent()) {
				continue
			}

			newLeaves = append(newLeaves, s.Parent())
		}
		leaves = newLeaves
	}

	return
}

// ─────────────────────────────────────────────────────────────────────────────
