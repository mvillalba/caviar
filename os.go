// os.go implements drop-in replacement functions for the os package.

package caviar

import (
    "os"
    "errors"
)

// Open mimicks os.Open. It will first attempt to open the file as an internal
// Caviar file and failing that it will pass along the call to the os package.
func Open(name string) (File, error) {
    file, err := CaviarOpen(name)
    if err != nil { return os.Open(name) }
    return file, nil
}

// OpenFile mimicks os.OpenFile. It will first attempt to open the file as an
// internal Caviar file and failing that it will pass along the call to the os
// package.
func OpenFile(name string, flag int, perm os.FileMode) (File, error) {
    file, err := CaviarOpenFile(name, flag, perm)
    if err != nil { return os.OpenFile(name, flag, perm) }
    return file, nil
}

// Lstat mimicks os.Lstat(). Please note Caviar does not currently support
// symlinks inside the bundle so Lstat() will be have identically to Stat() for
// paths matching files and directories inside the bundle.
func Lstat(name string) (fi os.FileInfo, err error) {
    return fi, debug(errors.New("Lstat(): Not Implemented.")) // TODO
}

// Stat mimicks os.Stat()
func Stat(name string) (fi os.FileInfo, err error) {
    return fi, debug(errors.New("Stat(): Not Implemented.")) // TODO
}
