package main

import (
	"os"
	"os/exec"

	"github.com/avamsi/climate"
	"github.com/avamsi/ergo/check"
)

func fileNames(dir string) []string {
	var files []string
	for _, d := range check.Ok(os.ReadDir(dir)) {
		if d.IsDir() {
			continue
		}
		files = append(files, d.Name())
	}
	return files
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

type folderphile struct{}

type diffOptions struct {
	left, right string
}

func (f *folderphile) Diff(opts *diffOptions) error {
	var (
		names = dedupeInOrder(fileNames(opts.left), fileNames(opts.right))
		files = make([]file, 0, len(names))
	)
	for _, name := range names {
		files = append(files, newFile("", opts.left, opts.right, "", name))
	}
	if tuiEditFiles(files, vsCodeDiff) {
		return nil
	}
	return climate.ErrExit(1)
}

type diff3Options struct {
	left, right, output string
}

func (f *folderphile) Diff3(opts *diff3Options) error {
	var (
		names = dedupeInOrder(fileNames(opts.left), fileNames(opts.right))
		files = make([]file, 0, len(names))
	)
	for _, name := range names {
		files = append(files, newFile(opts.right, opts.left, opts.right, opts.output, name))
	}
	if tuiEditFiles(files, vsCodeMerge) {
		return nil
	}
	return climate.ErrExit(1)
}

type mergeOptions struct {
	base, left, right, output string
}

func (f *folderphile) Merge(opts *mergeOptions) error {
	var (
		names = dedupeInOrder(fileNames(opts.left), fileNames(opts.right))
		files = make([]file, 0, len(names))
	)
	for _, name := range names {
		files = append(files, newFile(opts.base, opts.left, opts.right, opts.output, name))
	}
	if tuiEditFiles(files, vsCodeMerge) {
		return nil
	}
	return climate.ErrExit(1)
}

func main() {
	os.Exit(climate.Run(climate.Struct[folderphile]()))
}
