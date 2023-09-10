package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/avamsi/ergo/check"
)

type fileState int

const (
	fileStateUnknown fileState = iota
	fileUnedited
	fileSameAsBase
	fileSameAsLeft
	fileSameAsRight
	fileEditedOther
)

type file struct {
	base, left, right, output string // paths (empty if not applicable)
	relpath                   string
	state                     fileState

	baseContent, leftContent, rightContent []byte
	initialContent                         []byte
}

func readFile(path string) []byte {
	return check.Ok(os.ReadFile(path))
}

func newFile(base, left, right, output, relpath string) file {
	f := file{
		left:    filepath.Join(left, relpath),
		right:   filepath.Join(right, relpath),
		relpath: relpath,
		state:   fileUnedited,
	}
	f.leftContent = readFile(f.left)
	f.rightContent = readFile(f.right)
	if base != "" {
		f.base = filepath.Join(base, relpath)
		f.baseContent = readFile(f.base)
	}
	if output != "" {
		f.output = filepath.Join(output, relpath)
		f.initialContent = readFile(f.output)
	} else {
		// Assumed to be a 2-way diff, where right is the editable side.
		check.Truef(base == "", "base is not empty (%q) in a 2-way diff", base)
		f.output = f.right
		f.initialContent = f.rightContent
	}
	return f
}

func (f *file) updateState() {
	switch content := readFile(f.output); {
	case bytes.Equal(content, f.initialContent):
		f.state = fileUnedited
	case bytes.Equal(content, f.leftContent):
		f.state = fileSameAsLeft
	case bytes.Equal(content, f.rightContent):
		f.state = fileSameAsRight
	// Same-as-base is not as likely as others, check last.
	case f.base != "" && bytes.Equal(content, f.baseContent):
		f.state = fileSameAsBase
	default:
		f.state = fileEditedOther
	}
}
