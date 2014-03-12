// file.go contains a drop-in replacement for the os.File API.

package caviar

import (
    "io"
    "os"
    "errors"
)

// Maximum number of bytes to read when calling Read().
const READ_MAX = 1024 * 32

// File mimicks the os.File type's entire public API so Caviar can serve as a
// drop-in replacement.
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

// CaviarFile implements caviar.File and serves as a replacement for os.File.
type CaviarFile struct {
    obj *Object
    fd  int64
    pos int64
}

// Read mimicks os.File.Read().
func (f *CaviarFile) Read(b []byte) (int, error) {
    // Directory? No can do!
    if f.obj.ModeBits.IsDir() {
        return 0, debug(errors.New("Can't read data from a directory."))
    }

    // How much are we going to read?
    n := int64(len(b))
    l := f.obj.Size - f.pos
    if n > READ_MAX { n = READ_MAX }
    if n > l { n = l }
    if n == 0 { return 0, io.EOF }

    // Make the copy
    data, err := getPayload(f.obj)
    if err != nil { return 0, debug(err) }

    copy(b, data[f.pos:f.pos+n])

    // Update read position
    f.pos += n

    return int(n), nil
}

// ReadAt mimicks os.File.ReadAt().
func (f *CaviarFile) ReadAt(b []byte, off int64) (int, error) {
    // Directory? No can do!
    if f.obj.ModeBits.IsDir() {
        return 0, debug(errors.New("Can't read data from a directory."))
    }

    // How much are we going to read?
    n := int64(len(b))
    l := f.obj.Size - off
    if n > l { n = l }
    if n == 0 { return 0, io.EOF }
    if n < 0 { return 0, errors.New("Can't read before the beginning of the file!") }

    // Make the copy
    data, err := getPayload(f.obj)
    if err != nil { return 0, debug(err) }

    copy(b, data[off:off+n])

    return int(n), nil
}

// Write mimicks os.File.Write(). It always returns an error as Caviar files
// are read-only.
func (f *CaviarFile) Write(b []byte) (n int, err error) {
    return 0, debug(errors.New("Can't write file: caviar files are read-only."))
}

// WriteAt mimicks os.File.WriteAt(). It always returns an error as Caviar
// files are read-only.
func (f *CaviarFile) WriteAt(b []byte, off int64) (int, error) {
    return 0, debug(errors.New("Can't write file: caviar files are read-only."))
}

// Seek mimicks os.File.Seek().
func (f *CaviarFile) Seek(offset int64, whence int) (pos int64, err error) {
    // Directory? No can do!
    if f.obj.ModeBits.IsDir() {
        return 0, debug(errors.New("Can't seek through a directory."))
    }

    // Seek, seek, seek!
    if        whence == os.SEEK_SET {
        pos = offset                // From start of file
    } else if whence == os.SEEK_CUR {
        pos = f.pos + offset        // From current position
    } else if whence == os.SEEK_END {
        pos = f.obj.Size + offset   // From end of file
    }

    // Did we go over or under?
    if f.obj.Size < pos {
        return f.pos, debug(errors.New("Attempted to Seek() beyond end of file."))
    } else if pos < 0 {
        return f.pos, debug(errors.New("Attempted to Seek() before start of file."))
    }

    f.pos = pos
    return pos, nil
}

// Close mimicks os.File.Close()
func (f *CaviarFile) Close() error {
    if f.obj == nil {
        return debug(errors.New("File already closed."))
    }
    f.obj = nil
    return nil
}

// Stat mimicks os.File.Stat().
func (f *CaviarFile) Stat() (os.FileInfo, error) {
    return &CaviarFileInfo{ f.obj }, nil
}

// Name mimicks os.File.Name().
func (f *CaviarFile) Name() string {
    return f.obj.Name
}

// Chdir mimicks os.File.Chdir(). It will return with an error unless
// EXTRACT_EXECUTABLE is used as the extraction mode for the bundle as it's not
// possible to chdir to a virtual, in-memory directory (EXTRACT_MEMORY) and
// doing so for EXTRACT_TEMP would screw up relative paths causing subtle bugs.
func (f *CaviarFile) Chdir() error {
    // TODO: account for extraction mode as the docstring explains.
    return debug(errors.New("Can't chdir to file's directory: caviar files exist only in memory."))
}

// Sync mimicks os.File.Sync(). It always returns an error as Caviar files are
// read-only.
func (f *CaviarFile) Sync() (err error) {
    return debug(errors.New("Can't sync file: caviar files are read-only."))
}

// Fd mimicks os.File.Fd(). The returned file descriptor is a dummy value that
// is unlikely to repeat across Open files (but no guarantees).
func (f *CaviarFile) Fd() uintptr {
    return uintptr(f.fd)
}

// Truncate mimicks os.File.Truncate(). It always returns an error as Caviar
// files are read-only.
func (f *CaviarFile) Truncate(size int64) error {
    return debug(errors.New("Can't truncate file: caviar files are read-only."))
}

// WriteString mimicks os.File.WriteString(). It always returns an error as
// Caviar files are read-only.
func (f *CaviarFile) WriteString(s string) (int, error) {
    if len(s) == 0 { return 0, nil }
    return 0, debug(errors.New("Can't write file: caviar files are read-only."))
}

// Chmod mimicks os.File.Chmod(). It always returns an error as Caviar files
// are read-only.
func (f *CaviarFile) Chmod(mode os.FileMode) error {
    return debug(errors.New("Can't chmod file: caviar files are read-only."))
}

// Chown mimicks os.File.Chown(). It always returns an error as Caviar files
// are read-only.
func (f *CaviarFile) Chown(uid, gid int) error {
    return debug(errors.New("Can't chown file: caviar files are read-only."))
}

// Readdir mimicks os.File.Readdir().
func (f *CaviarFile) Readdir(n int) (fi []os.FileInfo, err error) {
    // File? No can do!
    if !f.obj.ModeBits.IsDir() {
        return fi, debug(errors.New("Files can't contain other files and directories!."))
    }

    // Build dir list
    for i := f.pos; i < int64(len(f.obj.Objects)); i++ {
        if n > 0 && i == int64(n) { break }
        fi = append(fi, &CaviarFileInfo{ &f.obj.Objects[i] })
    }
    if n > 0 && len(fi) == 0 { return fi, io.EOF }

    f.pos += int64(len(fi))
    return fi, nil
}

// Readdirnames mimicks os.File.Readdirnames().
func (f *CaviarFile) Readdirnames(n int) (names []string, err error) {
    // File? No can do!
    if !f.obj.ModeBits.IsDir() {
        return names, debug(errors.New("Files can't contain other files and directories!."))
    }

    // Build dir list
    for i := f.pos; i < int64(len(f.obj.Objects)); i++ {
        if n > 0 && i == int64(n) { break }
        names = append(names, f.obj.Objects[i].Name)
    }
    if n > 0 && len(names) == 0 { return names, io.EOF }

    f.pos += int64(len(names))
    return names, nil
}
