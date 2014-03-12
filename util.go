// util.go implements various auxiliary functions.

package caviar

import (
    "log"
    "errors"
    "strings"
    "time"
    "os"
    "path"
    "path/filepath"
)

// File extension for Caviar containers.
const CAVIAR_EXTENSION = "cvr"

// Returns the name of a detached container for a given program “p”.
func DetachedName(p string) string {
    ext := path.Ext(p)
    if ext != "" {
        p = p[:len(p)-len(ext)-1]
    }
    return p + "." + CAVIAR_EXTENSION
}

// Returns the total number of bytes for all loaded assets.
func PayloadSize() int64 {
    return int64(len(state.assets))
}

// Given a file Object, return a (zero-copy) slice containing the object's data.
func getPayload(obj *Object) ([]byte, error) {
    if obj.ModeBits.IsDir() {
        return nil, debug(errors.New("Directories have no payload!"))
    }
    if obj.Size == 0 {
        return nil, debug(errors.New("The file is empty!"))
    }
    return state.assets[obj.Offset:obj.Offset+obj.Size], nil
}

// Self-explanatory debug helpers.
func isDebug() bool {
    return state.manifest.Options.Debug
}

func debug(v interface{}) error {
    if v == nil { return nil }
    if isDebug() {
        log.Print("[CAVIAR] ", v)
    }
    if err, ok := v.(error); ok { return err }
    if err, ok := v.(string); ok { return errors.New(err) }
    return errors.New("debug(): Got message of unknown type.")
}

// CaviarOpen behaves the same way as Open but it will only attempt to open
// files and directories contained within the bundle and will not pass along
// the call to os.Open() on failure.
func CaviarOpen(name string) (File, error) {
    return CaviarOpenFile(name, 0, 0)
}

// CaviarOpenFile is to OpenFile what CaviarOpen is to Open.
func CaviarOpenFile(name string, flag int, perm os.FileMode) (File, error) {
    if !state.ready { return nil, debug(errors.New("Caviar is not ready.")) }

    // TODO: Validate flags and permissions (can't open a Caviar file for
    // writing, after all).

    obj, err := findObject(name)
    if err != nil { return nil, debug(err) }

    return &CaviarFile{ obj, genFd(obj), 0 }, nil
}

// Given a path, find the corresponding Object. Returns an error if not found.
func findObject(name string) (obj *Object, err error) {
    // TODO: Handle volumes names and implement case-insensitive matches for
    // Windows support.

    // Turn relative paths to absolute paths
    if !path.IsAbs(name) {
        name, err = filepath.Abs(filepath.Clean(name))
        if err != nil { return nil, debug(err) }
    }

    // Does the path refer to the object root specifically?
    if name == state.prefix {
        return &state.manifest.ObjectRoot, nil
    }

    // Does it even point to a file inside the object root?
    if !strings.HasPrefix(name, state.prefix) {
        return nil, debug(errors.New("Caviar file not found: " + name))
    }

    // Turn absolute path into a relative path within the object root
    name = name[len(state.prefix)+1:]

    // Find object
    segments := strings.Split(name, string(os.PathSeparator))
    curobj := &state.manifest.ObjectRoot
    for _, segment := range segments {
        match := false
        for i, o := range curobj.Objects {
            if o.Name == segment {
                curobj = &curobj.Objects[i]
                match = true
                break
            }
        }
        if !match {
            return nil, debug(errors.New("Caviar file not found: " + name))
        }
    }

    return curobj, nil
}

// Generate a probably unique FD.
func genFd(obj *Object) int64 {
    return int64(obj.Checksum) + time.Now().Unix()
}

