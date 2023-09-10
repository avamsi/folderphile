package main

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	_ "embed"

	"github.com/avamsi/climate"
	"github.com/avamsi/ergo/check"
)

func fileRelpaths(dir string) (paths []string) {
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		paths = append(paths, check.Ok(filepath.Rel(dir, path)))
		return nil
	}
	check.Nil(filepath.WalkDir(dir, walk))
	return paths
}

// TODO: remove after go1.21.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func dedupeInOrder(s1, s2 []string) []string {
	out := make([]string, 0, max(len(s1), len(s2)))
	for {
		if len(s1) == 0 || len(s2) == 0 {
			out = append(out, s1...)
			out = append(out, s2...)
			return out
		}
		if s1[0] == s2[0] {
			out = append(out, s1[0])
			s1, s2 = s1[1:], s2[1:]
		} else if s1[0] < s2[0] {
			out = append(out, s1[0])
			s1 = s1[1:]
		} else {
			out = append(out, s2[0])
			s2 = s2[1:]
		}
	}
}

var (
	vsCodeDiff = func(f *file) *exec.Cmd {
		return exec.Command(
			"code",
			"--diff", f.left, f.right,
			"--wait", "--new-window",
		)
	}
	vsCodeMerge = func(f *file) *exec.Cmd {
		return exec.Command(
			"code",
			"--merge", f.left, f.right, f.base, f.output,
			"--wait", "--new-window",
		)
	}
)

type options struct {
	// base is the common ancestor of left and right; implies merge
	base        string `climate:"short"`
	left, right string `climate:"short,required"` // side to compare
	output      string `climate:"short"`          // output is the destination
}

// folderphile is a diff / merge editor (depending on whether "base" is set)
// that recursively compares two folders ("left" and "right").
func folderphile(opts *options) error {
	var (
		base string
		e    = vsCodeMerge
	)
	switch {
	case opts.base == "" && opts.output == "":
		// Assumed to be a 2-way diff.
		e = vsCodeDiff
	case opts.base == "" && opts.output != "":
		// Assumed to be a 3-way diff, where left is the base.
		base = opts.left
	case opts.base != "" && opts.output == "":
		err := errors.New(`required flag "output" not set (required with "base")`)
		return climate.ErrUsage(err)
	case opts.base != "" && opts.output != "":
		// Assumed to be a merge (since both base and output are not empty).
		base = opts.base
	}
	var (
		paths = dedupeInOrder(fileRelpaths(opts.left), fileRelpaths(opts.right))
		files = make([]file, 0, len(paths))
	)
	for _, rel := range paths {
		files = append(files, newFile(base, opts.left, opts.right, opts.output, rel))
	}
	if tuiEditFiles(files, e) {
		return nil
	}
	return climate.ErrExit(1)
}

//go:generate go run github.com/avamsi/climate/cmd/climate --out=md.climate
//go:embed md.climate
var md []byte

func main() {
	os.Exit(climate.Run(climate.Func(folderphile), climate.Metadata(md)))
}
