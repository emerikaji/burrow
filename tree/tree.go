package tree

import (
	"strings"

	"github.com/emerikaji/burrow/index"
)

type Entry struct {
	Hash     [20]byte
	Children map[string]*Entry
}

func Build(idx *index.Index) (tree *Entry, err error) {
	tree = &Entry{
		Children: make(map[string]*Entry),
	}

	for _, indexEntry := range *idx {
		pathSplit := strings.Split(indexEntry.Path, "/")

		treeLocation := tree
		for i, name := range pathSplit {
			if i == len(pathSplit)-1 {
				treeLocation.Children[name] = &Entry{
					Hash: indexEntry.Hash,
				}
			}

			if _, ok := treeLocation.Children[name]; !ok {
				treeLocation.Children[name] = &Entry{
					Children: make(map[string]*Entry),
				}
			}

			treeLocation = treeLocation.Children[name]
		}
	}

	return
}
