// init.go handles the runtime initialization of Caviar.

package caviar

import (
    "bitbucket.org/kardianos/osext"
    "archive/zip"
    "encoding/gob"
    "io"
    "io/ioutil"
    "errors"
    "path"
    "fmt"
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
    if state.ready { return debug(errors.New("Already initialized.")) }

    // Setup global state
    state.prefix, err = osext.Executable()
    if err != nil { return debug(err) }
    state.prefix = path.Dir(state.prefix)

    // Load ZIP container
    path, err := osext.Executable()
    if err != nil { return debug(err) }

    reader, err := zip.OpenReader(path)
    if err != nil {
        reader, err = zip.OpenReader(DetachedName(path))
        if err != nil { return debug(err) }
    }
    defer reader.Close()

    // Load manifest
    m, err := getFile(reader, "Manifest.gob")
    if err != nil { return debug(err) }

    dec := gob.NewDecoder(m)
    err = dec.Decode(&state.manifest)
    if err != nil { return debug(err) }

    // Process bundle options
    if state.manifest.Options.CustomPrefix != "" {
        state.prefix = state.manifest.Options.CustomPrefix
    }

    if state.manifest.Options.ExtractionMode != EXTRACT_MEMORY {
        return debug(errors.New("Unsupported extraction mode: only EXTRACT_MEMORY is currently supported."))
    }

    // Load assets
    // TODO: account for extraction mode
    a, err := getFile(reader, "Assets.bin")
    if err != nil { return debug(err) }

    state.assets, err = ioutil.ReadAll(a)
    if err != nil {
        state.assets = nil
        return debug(err)
    }

    // Verify manifest
    err = verifyManifest()
    if err != nil {
        state.assets = nil
        return debug(err)
    }

    // All done
    debug("Caviar is ready.")
    debug(fmt.Sprintf("Loaded %v bytes.", PayloadSize()))
    state.ready = true
    return nil
}

// Find file inside a ZIP container.
func getFile(reader *zip.ReadCloser, name string) (io.Reader, error) {
    for _, f := range reader.File {
        if f.Name != name { continue }
        r, err := f.Open()
        return r, debug(err)
    }
    return nil, debug(errors.New("File not found: " + name))
}

// Call Init() automatically on startup.
func init() {
    Init()
}
