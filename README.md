```
$ go install github.com/avamsi/folderphile@latest
```

```
$ folderphile --help

folderphile is a diff / merge editor (depending on whether "base" is set)
that recursively compares two folders ("left" and "right").

Usage:
  folderphile [opts]

Flags:
  -b, --base    string   base is the common ancestor of left and right; implies merge
  -l, --left    string   side to compare
  -r, --right   string   side to compare
  -o, --output  string   output is the destination
  -h, --help             help for folderphile
  -v, --version          version for folderphile
```
