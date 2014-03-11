package caviar

import (
    "io"
    "os"
    "strings"
    "path"
    "errors"
    "time"
)

const READ_MAX = 1024 * 32 // 32 KiB

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

func caviarOpen(name string, flag int, perm os.FileMode, short bool) (f File, err error) {
    // Locking for the greater good
    state.mu.Lock()
    defer state.mu.Unlock()

    found := false
    npath := name
    var file *fileInfo
    var dirinfo *directoryInfo

    // Turn relative paths to absolute paths
    if !strings.HasPrefix(npath, "/") {
        cwd, err := os.Getwd()
        if err != nil { return f, err }
        npath = path.Join(cwd, npath)
    }

    if strings.HasPrefix(npath, assetpath) {
        // Convert matching absolute paths back to relative paths
        npath = npath[len(assetpath):]

        // Attempt to resolve path down to a fileInfo instance
        file, dirinfo, err = findAsset(name, npath)
        if err != nil {
            found = false
        } else {
            found = true
        }
    }

    if !found {
        if short {
            return os.Open(name)
        } else {
            return os.OpenFile(name, flag, perm)
        }
    }

    fd := mkfd()
    state.descriptors[fd] = new(fileDescriptor)
    state.descriptors[fd].file = file
    state.descriptors[fd].directory = dirinfo
    state.descriptors[fd].position = 0 // Just in case.

    return &CaviarFile{ uint(fd) }, nil
}

// Call this function after acquiring the global state mutex!
func mkfd() int {
    for i, entry := range state.descriptors {
        if entry == nil { return i }
    }
    state.descriptors = append(state.descriptors, nil)
    return len(state.descriptors) - 1
}

// Call this function after acquiring the global state mutex!
func findAsset(name, npath string) (file *fileInfo, dirinfo *directoryInfo, err error) {
    segments := strings.Split(npath, "/")
    fileordir := segments[len(segments)-1]
    segments = segments[:len(segments)-1]

    // TODO: should handle opening the asset root directly and merge with os
    // output if needed.

    // Find directory
    curdir := &state.root
    for _, segment := range segments {
        if segment == "." { continue }
        if segment == ".." {
            if curdir.parent != nil {
                curdir = curdir.parent
                continue
            } else {
                return file, dirinfo, errors.New("Caviar file not found: " + name)
            }
        }
        for i, dir := range curdir.directories {
            if dir.name == segment {
                curdir = &curdir.directories[i]
                continue
            }
        }
        return file, dirinfo, errors.New("Caviar file not found: " + name)
    }

    // Is it a file?
    for i, entry := range curdir.files {
        if entry.name == fileordir {
            return &curdir.files[i], nil, nil
        }
    }

    // Is it a directory?
    for i, entry := range curdir.directories {
        // TODO: handle '.' and '..'?
        if entry.name == fileordir {
            return nil, &curdir.directories[i], nil
        }
    }

    return file, dirinfo, errors.New("Caviar file not found: " + name)
}

func Open(name string) (file File, err error) {
    if !cavinit { Init() }
    if bypass { return os.Open(name) }
    return caviarOpen(name, 0, 0, true)
}

func OpenFile(name string, flag int, perm os.FileMode) (file File, err error) {
    if !cavinit { Init() }
    if bypass { return os.OpenFile(name, flag, perm) }
    return caviarOpen(name, flag, perm, false)
}

// io.Reader
func (f *CaviarFile) Read(b []byte) (n int, err error) {
    // How to avoid this…?
    state.mu.Lock()
    defer state.mu.Unlock()

    // Is it a dir?
    if state.descriptors[f.fd].file == nil {
        return 0, errors.New("Can't read data from a directory.")
    }

    // How much are we going to read?
    n = len(b)
    l := int(state.descriptors[f.fd].file.size - state.descriptors[f.fd].position)
    if n > READ_MAX { n = READ_MAX }
    if n > l { n = l }
    if n == 0 { return 0, io.EOF }

    // Make the copy
    start := state.descriptors[f.fd].file.offset + state.descriptors[f.fd].position
    end := int(start) + n
    copy(b, state.assets[start:end])

    // Update FD position
    state.descriptors[f.fd].position += uint(n)

    return n, nil
}

// io.ReaderAt
func (f *CaviarFile) ReadAt(b []byte, off int64) (n int, err error) {
    return n, errors.New("Not Implemented: ReadAt()") // TODO
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
    // Lock'em up
    state.mu.Lock()
    defer state.mu.Unlock()

    // Is it a dir?
    if state.descriptors[f.fd].file == nil {
        return 0, errors.New("Can't seek through a directory.")
    }

    if whence == 0 {
        state.descriptors[f.fd].position = uint(offset)
    } else if whence == 1 {
        state.descriptors[f.fd].position += uint(offset)
        if state.descriptors[f.fd].file.size < state.descriptors[f.fd].position {
            state.descriptors[f.fd].position = state.descriptors[f.fd].file.size
        }
    } else if whence == 2 {
        state.descriptors[f.fd].position = uint(offset) + state.descriptors[f.fd].file.size
    }

    return int64(state.descriptors[f.fd].position), nil
}

// io.Closer
func (f *CaviarFile) Close() error {
    state.mu.Lock()
    defer state.mu.Unlock()

    if state.descriptors[f.fd] == nil {
        return errors.New("File already closed.")
    }

    state.descriptors[f.fd] = nil
    return nil
}

// os.FileInfo
type CaviarFileInfo struct {
    file        *fileInfo
    directory   *directoryInfo
}

func (fi *CaviarFileInfo) Name() string {
    // Files are read-only and untouched after Init() returns so no locking is
    // needed :-).
    if fi.file == nil {
        return fi.directory.name
    }
    return fi.file.name
}

func (fi *CaviarFileInfo) Size() int64 {
    // Files are read-only and untouched after Init() returns so no locking is
    // needed :-).
    if fi.file == nil {
        return 0
    }
    return int64(fi.file.size)
}

func (fi *CaviarFileInfo) Mode() os.FileMode {
    chmod := os.FileMode(0664)
    if fi.file == nil {
        return os.ModeDir | chmod
    }
    return chmod
}

func (fi *CaviarFileInfo) ModTime() time.Time {
    return modtime
}

func (fi *CaviarFileInfo) IsDir() bool {
    return fi.Mode().IsDir()
}

func (fi *CaviarFileInfo) Sys() interface{} {
    return nil
}

// os.File
func (f *CaviarFile) Stat() (fi os.FileInfo, err error) {
    state.mu.Lock()
    defer state.mu.Unlock()

    return &CaviarFileInfo{ state.descriptors[f.fd].file,
                            state.descriptors[f.fd].directory }, nil
}

func (f *CaviarFile) Name() string {
    state.mu.Lock()
    defer state.mu.Unlock()

    if state.descriptors[f.fd].file == nil {
        return state.descriptors[f.fd].directory.name
    }
    return state.descriptors[f.fd].file.name
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
    return fi, errors.New("Not Implemented: Readdir().") // TODO
}

func (f *CaviarFile) Readdirnames(n int) (names []string, err error) {
    return names, errors.New("Not Implemented: Readdir().") // TODO
}
