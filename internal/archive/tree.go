package archive

import (
	"fmt"
	"io"
)

// Tree drawing characters (Unicode box-drawing).
const (
	treeBranch = "├── "
	treeLast   = "└── "
	treeVert   = "│   "
	treeSpace  = "    "
)

// TreePrinter renders directory trees with Unicode box characters.
type TreePrinter struct {
	DirsFirst bool
	Writer    io.Writer
}

// Print renders the tree starting from root.
// The root entry itself is printed, followed by its children.
func (p *TreePrinter) Print(root *DirEntry) {
	// Print root name
	name := root.Name
	if root.IsDir {
		name += "/"
	}
	fmt.Fprintln(p.Writer, name)

	// Print children
	p.printChildren(root.Children, "")
}

func (p *TreePrinter) printChildren(children []*DirEntry, prefix string) {
	if p.DirsFirst {
		SortDirsFirst(children)
	}

	for i, child := range children {
		isLast := i == len(children)-1

		// Choose the appropriate connector
		var connector string
		if isLast {
			connector = treeLast
		} else {
			connector = treeBranch
		}

		// Format the entry name
		name := child.Name
		if child.IsDir {
			name += "/"
		}

		// Print this entry
		fmt.Fprintf(p.Writer, "%s%s%s\n", prefix, connector, name)

		// If this is a directory with children, recurse
		if child.IsDir && len(child.Children) > 0 {
			var childPrefix string
			if isLast {
				childPrefix = prefix + treeSpace
			} else {
				childPrefix = prefix + treeVert
			}
			p.printChildren(child.Children, childPrefix)
		}
	}
}

// Counts returns the number of directories and files in a tree.
// The root directory is not counted.
func Counts(root *DirEntry) (dirs, files int) {
	countRecursive(root, &dirs, &files)
	return dirs, files
}

func countRecursive(entry *DirEntry, dirs, files *int) {
	for _, child := range entry.Children {
		if child.IsDir {
			*dirs++
			countRecursive(child, dirs, files)
		} else {
			*files++
		}
	}
}
