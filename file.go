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
	name                      string
	state                     fileState

	baseContent, leftContent, rightContent []byte
	initialContent                         []byte
}

func readFile(path string) []byte {
	return check.Ok(os.ReadFile(path))
}

func newFile(base, left, right, output, name string) file {
	f := file{
		left:  filepath.Join(left, name),
		right: filepath.Join(right, name),
		name:  name,
		state: fileUnedited,
	}
	f.leftContent = readFile(f.left)
	f.rightContent = readFile(f.right)
	if base != "" {
		f.base = filepath.Join(base, name)
		f.baseContent = readFile(f.base)
	}
	if output != "" {
		f.output = filepath.Join(output, name)
		f.initialContent = readFile(f.output)
	} else {
		// Right is assumed to be the editable side when output is not
		// explicitly specified (diff mode, for example).
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
	// Same as base is not as likely as others, check last.
	case bytes.Equal(content, f.baseContent):
		f.state = fileSameAsBase
	default:
		f.state = fileEditedOther
	}
}
