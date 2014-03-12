// fileinfo.go implements the os.FileInfo interface for Caviar files.

package caviar

import (
    "time"
    "os"
)

// CaviarFileInfo implements os.FileInfo and should behave exactly the same as
// the os package's implementation for native OS files.
type CaviarFileInfo struct {
    obj *Object
}

func (fi *CaviarFileInfo) Name() string {
    return fi.obj.Name
}

func (fi *CaviarFileInfo) Size() int64 {
    return fi.obj.Size
}

func (fi *CaviarFileInfo) Mode() os.FileMode {
    return fi.obj.ModeBits
}

func (fi *CaviarFileInfo) ModTime() time.Time {
    return time.Unix(fi.obj.ModTime, 0)
}

func (fi *CaviarFileInfo) IsDir() bool {
    return fi.Mode().IsDir()
}

func (fi *CaviarFileInfo) Sys() interface{} {
    return nil
}
