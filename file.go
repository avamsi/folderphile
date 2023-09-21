package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/avamsi/ergo/assert"
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

func readFile(path string) ([]byte, bool) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false
	}
	assert.Nil(err)
	return b, true
}

func newFile(base, left, right, output, relpath string) file {
	f := file{
		left:    filepath.Join(left, relpath),
		right:   filepath.Join(right, relpath),
		relpath: relpath,
		state:   fileUnedited,
	}
	var ok bool
	if f.leftContent, ok = readFile(f.left); !ok {
		f.left = "/dev/null"
	}
	f.rightContent, ok = readFile(f.right)
	// TODO: we require right to exist for now, but this is not ideal as it's
	// possible for right to be deleted. folderphile should gracefully allow the
	// user to create right again, instead of just panicking.
	assert.Truef(ok, "right file %q does not exist", f.right)
	if base != "" {
		f.base = filepath.Join(base, relpath)
		if f.baseContent, ok = readFile(f.base); !ok {
			f.base = "/dev/null"
		}
	}
	if output != "" {
		f.output = filepath.Join(output, relpath)
		f.initialContent, ok = readFile(f.output)
		// TODO: similar to right above, output need not exist.
		assert.Truef(ok, "output file %q does not exist", f.output)
	} else {
		// Assumed to be a 2-way diff, where right is the editable side.
		assert.Truef(base == "", "base is not empty (%q) in a 2-way diff", base)
		f.output = f.right
		f.initialContent = f.rightContent
	}
	return f
}

func (f *file) updateState() {
	switch content, _ := readFile(f.output); {
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
