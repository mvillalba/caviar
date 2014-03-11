package caviar

import (
    "bitbucket.org/kardianos/osext"
    "archive/zip"
    "fmt"
    "sync"
    "encoding/gob"
    "io"
    "io/ioutil"
    "errors"
    "time"
)

type caviarState struct {
    mu          sync.Mutex
    root        directoryInfo
    descriptors []*fileDescriptor
    assets      []byte
}

type fileInfo struct {
    name        string
    parent      *directoryInfo
    offset      uint
    size        uint
}

type directoryInfo struct {
    name        string
    parent      *directoryInfo
    files       []fileInfo
    directories []directoryInfo
}

type fileDescriptor struct {
    file        *fileInfo
    directory   *directoryInfo
    position    uint
}

const DESCRIPTOR_FLOOR = 42000000

// Global state
var state caviarState
var cavinit = false
var bypass = true
var modtime = time.Now()
var assetpath = ""

// Where all the magic happens
func Init() (int, error) {
    // Mutexes for the greater good.
    state.mu.Lock()
    defer state.mu.Unlock()

    // Setup global state
    var err error
    cavinit = true
    assetpath, err = osext.ExecutableFolder()
    if err != nil { return len(state.assets), err }

    // Load ZIP container
    path, err := osext.Executable()
    if err != nil { return len(state.assets), err }

    reader, err := zip.OpenReader(path)
    if err != nil {
        reader, err = zip.OpenReader(DetachedName(path))
        if err != nil { return len(state.assets), err }
    }
    defer reader.Close()

    // Load manifest
    var manifest Manifest

    m, err := getFile(reader, "Manifest.gob")
    if err != nil { return len(state.assets), err }

    dec := gob.NewDecoder(m)
    err = dec.Decode(&manifest)
    if err != nil { return len(state.assets), err }

    // Load assets
    a, err := getFile(reader, "Assets.bin")
    if err != nil { return len(state.assets), err }

    state.assets, err = ioutil.ReadAll(a)
    if err != nil { return len(state.assets), err }

    // Process manifest
    count, err := processManifest(&manifest, &state.root)
    if err != nil { return len(state.assets), err }
    // BEGIN HACK
    state.root.files = state.root.directories[0].files
    state.root.directories = state.root.directories[0].directories
    // END HACK

    if count != len(state.assets) {
        errstr := "Asset payload size (%v) differs from manifest tally (%v)."
        errstr += " Something is really, really wrong."
        errstr = fmt.Sprintf(errstr, len(state.assets), count)
        return len(state.assets), errors.New(errstr)
    }

    bypass = false
    return len(state.assets), nil
}

func getFile(reader *zip.ReadCloser, name string) (io.Reader, error) {
    for _, f := range reader.File {
        if f.Name != name { continue }
        return f.Open()
    }
    return nil, errors.New("File not found: " + name)
}

