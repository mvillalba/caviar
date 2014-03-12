// init.go handles the runtime initialization of Caviar.

package caviar

import (
    "bitbucket.org/kardianos/osext"
    "archive/zip"
    "encoding/gob"
    "io"
    "io/ioutil"
    "errors"
    "log"
)

// Global state
type caviarState struct {
    manifest    Manifest
    assets      []byte
    ready       bool
    prefix      string
}

var state caviarState

// Init sets up Caviar's internal state and loads the bundle, if any.
func Init() (err error) {
    // Setup global state
    state.ready = false
    state.prefix, err = osext.ExecutableFolder()
    if err != nil { return err }

    // Load ZIP container
    path, err := osext.Executable()
    if err != nil { return err }

    reader, err := zip.OpenReader(path)
    if err != nil {
        reader, err = zip.OpenReader(DetachedName(path))
        if err != nil { return err }
    }
    defer reader.Close()

    // Load manifest
    m, err := getFile(reader, "Manifest.gob")
    if err != nil { return err }

    dec := gob.NewDecoder(m)
    err = dec.Decode(&state.manifest)
    if err != nil { return err }

    // Process bundle options
    if state.manifest.Options.CustomPrefix != "" {
        state.prefix = state.manifest.Options.CustomPrefix
    }

    // Load assets
    // TODO: account for extraction mode
    a, err := getFile(reader, "Assets.bin")
    if err != nil { return err }

    state.assets, err = ioutil.ReadAll(a)
    if err != nil { return err }

    // Verify manifest
    err = verifyManifest()
    if err != nil { return err }

    // All done
    debug("Caviar is ready.")
    state.ready = true
    return nil
}

// Find file inside a ZIP container.
func getFile(reader *zip.ReadCloser, name string) (io.Reader, error) {
    for _, f := range reader.File {
        if f.Name != name { continue }
        return f.Open()
    }
    return nil, errors.New("File not found: " + name)
}

// Call Init() automatically on startup.
func init() {
    err := Init()
    if err != nil && isDebug() {
        log.Fatal(err)
    }
}
