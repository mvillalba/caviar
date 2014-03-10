package caviar

import (
    "io"
    "os"
    "errors"
)

// In-memory file
type CaviarFile struct {
    fd      uint
}

// os.File and friends mocking
type File interface {
    io.Reader
    io.ReaderAt
    io.Writer
    io.WriterAt
    io.Seeker
    io.Closer
    Stat() (os.FileInfo, error)
    Name() string
    Chdir() error
    Sync() error
    Fd() uintptr
    Truncate(int64) error
    WriteString(string) (int, error)
    Chmod(os.FileMode) error
    Chown(int, int) error
    Readdir(n int) (fi []os.FileInfo, err error)
    Readdirnames(n int) (names []string, err error)
}

func Open(name string) (file File, err error) {
    return &CaviarFile{}, nil // TODO
}

func OpenFile(name string, flag int, perm os.FileMode) (file File, err error) {
    return os.OpenFile(name, flag, perm) // TODO
}

// io.Reader
func (f *CaviarFile) Read(b []byte) (n int, err error) {
    return n, nil // TODO
}

// io.ReaderAt
func (f *CaviarFile) ReadAt(b []byte, off int64) (n int, err error) {
    return n, nil // TODO
}

// io.Writer
func (f *CaviarFile) Write(b []byte) (n int, err error) {
    if len(b) == 0 { return 0, nil }
    return 0, errors.New("Can't write file: caviar files are read-only.")
}

// io.WriterAt
func (f *CaviarFile) WriteAt(b []byte, off int64) (int, error) {
    if len(b) == 0 { return 0, nil }
    return 0, errors.New("Can't write file: caviar files are read-only.")
}

// io.Seeker
func (f *CaviarFile) Seek(offset int64, whence int) (ret int64, err error) {
    return ret, nil // TODO
}

// io.Closer
func (f *CaviarFile) Close() error {
    return nil // TODO
}

// os.File
func (f *CaviarFile) Stat() (fi os.FileInfo, err error) {
    return fi, nil // TODO
}

func (f *CaviarFile) Name() string {
    return "caviar_is_yummy.go" // TODO
}

func (f *CaviarFile) Chdir() error {
    return errors.New("Can't chdir to file's directory: caviar files exist only in memory.")
}

func (f *CaviarFile) Sync() (err error) {
    return errors.New("Can't sync file: caviar files are read-only.")
}

func (f *CaviarFile) Fd() uintptr {
    return uintptr(f.fd)
}

func (f *CaviarFile) Truncate(size int64) error {
    return errors.New("Can't truncate file: caviar files are read-only.")
}

func (f *CaviarFile) WriteString(s string) (int, error) {
    if len(s) == 0 { return 0, nil }
    return 0, errors.New("Can't write file: caviar files are read-only.")
}

func (f *CaviarFile) Chmod(mode os.FileMode) error {
    return errors.New("Can't chmod file: caviar files are read-only.")
}

func (f *CaviarFile) Chown(uid, gid int) error {
    return errors.New("Can't chown file: caviar files are read-only.")
}

func (f *CaviarFile) Readdir(n int) (fi []os.FileInfo, err error) {
    return fi, nil // TODO
}

func (f *CaviarFile) Readdirnames(n int) (names []string, err error) {
    return names, nil // TODO
}
