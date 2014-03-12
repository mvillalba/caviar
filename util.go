// util.go implements various auxiliary functions.

package caviar

import (
    "path"
    "log"
    "errors"
    "strings"
    "time"
    "os"
)

// File extension for Caviar containers.
const CAVIAR_EXTENSION = "cvr"

// Returns the name of a detached container for a given program “p”.
func DetachedName(p string) string {
    ext := path.Ext(p)
    if ext != "" {
        p = p[:len(p)-len(ext)] + CAVIAR_EXTENSION
    }
    return p
}

// Returns the total number of bytes for all loaded assets.
func PayloadSize() int64 {
    return int64(len(state.assets))
}

// Given a file Object, return a (zero-copy) slice containing the object's data.
func getPayload(obj *Object) ([]byte, error) {
    if obj.ModeBits.IsDir() {
        return nil, errors.New("Directories have no payload!")
    }
    if obj.Size == 0 {
        return nil, errors.New("The file is empty!")
    }
    return state.assets[obj.Offset:obj.Offset+obj.Size], nil
}

// Self-explanatory debug helpers.
func isDebug() bool {
    return state.manifest.Options.Debug
}

func debug(message string) {
    if isDebug() {
        log.Print(message)
    }
}

//
func caviarOpen(name string, flag int, perm os.FileMode) (f File, err error) {
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

// Generate a probably unique FD.
func genFd(obj *Object) int64 {
    return obj.Checksum + time.Now().Unix()
}

