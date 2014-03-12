// manifest.go implements the foundational Manifest and Object types as well as
// some helper functions to verify their integrity.

// Caviar is an asset/resource packer for Go.
package caviar

import (
    "fmt"
    "errors"
    "os"
    "hash/crc32"
)

// Manifest-level magic value
const MANIFEST_MAGIC = "CAVIAR"
// Asset-root-level magic value
const OBJECTROOT_MAGIC = "ROOT"
const (
    // Load everything to RAM
    EXTRACT_MEMORY      = iota
    // Unpack everything to a temp dir and proxy OS Open/OpenFile calls to make
    // the resources available from the expected FS locations. Files are
    // deleted cleanly when the program exits.
    EXTRACT_TEMP
    // Same as EXTRACT_TEMP, but extract to the executable's root directory.
    EXTRACT_EXECUTABLE
)

// Manifest describes the contents of a Caviar bundle and it's serialized by
// cavundle along with the raw data of all assets when creating a bundle.
type Manifest struct {
    // Magic should be set to MANIFEST_MAGIC
    Magic           string
    // Free-formatted comment string added by the program generating the
    // bundle. Caviar's own cavundle adds information about Caviar and a
    // copyright statement.
    Comment         string
    // Additional options set when creating the bundle. These are needed as
    // Caviar's initialization routine is called from within an init() method
    // and thus can take no input from the program itself.
    Options         BundleOptions
    // The root directory object.
    ObjectRoot      Object
}

// Various options to be set by the program creating the bundle. They will
// affect Caviar's run-time behaviour.
type BundleOptions struct {
    // The asset root will be placed in a fixed location (i.e.
    // “/usr/share/myprogram”) instead of the executable's directory, whatever
    // that happens to be if CustomPrefix is set to anything other than an
    // empty string.
    CustomPrefix    string
    // This will make Caviar print out various log messages to the terminal.
    Debug           bool
    // See EXTRACT_* constants above.
    ExtractionMode  int
}

// Object represents either a file or a directory inside the bundle.
type Object struct {
    // Name of the file or directory. This is NOT a full path but rather one
    // segment (for instance, the file /home/martin/DATA would require 3
    // objects named “home”, “martin”, and “DATA”).
    Name        string
    // FileMode as returned by os.File.Stat(). Note that os.ModeDir must be set
    // if the object represents a directory.
    ModeBits    os.FileMode
    // Modification time.
    ModTime     int64
    // Size of the file. Set to 0 for directories.
    Size        int64
    // All assets inside a bundle are packed inside a single flat binary file
    // next to the serialized Manifest. Offset represents the offset from the
    // beginning of the asset binary file to the beginning of the data for the
    // file the given Object represents. Must be set to 0 for directories and
    // empty files.
    Offset      int64
    // CRC32 checksum for the file's contents. Set to 0 for directories.
    Checksum    uint32
    // Child objects (sub-directories and contained files). File objects must
    // not have any children.
    Objects     []Object
}

// Perform various sanity checks on the manifest and contained object tree.
func verifyManifest() (error) {
    // Verify magic
    if state.manifest.Magic != MANIFEST_MAGIC {
        errstr := "Container has invalid magic value (expected %v, got %v)."
        errstr = fmt.Sprintf(errstr, MANIFEST_MAGIC, state.manifest.Magic)
        return debug(errors.New(errstr))
    }

    // Verify asset root name tag magic
    if state.manifest.ObjectRoot.Name != OBJECTROOT_MAGIC {
        errstr := "Container has invalid magic value (expected %v, got %v)."
        errstr = fmt.Sprintf(errstr, OBJECTROOT_MAGIC, state.manifest.ObjectRoot.Name)
        return debug(errors.New(errstr))
    }

    // Verify options
    emode := state.manifest.Options.ExtractionMode
    if emode != EXTRACT_MEMORY && emode != EXTRACT_TEMP && emode != EXTRACT_EXECUTABLE {
        return debug(errors.New("Bundle specifies unknown extraction mode."))
    }

    // Verify object tree
    if !state.manifest.ObjectRoot.ModeBits.IsDir() {
        return debug(errors.New("Root Object must be a directory."))
    }

    count, err := verifyObject(&state.manifest.ObjectRoot)
    if err != nil { return debug(err) }

    // Verify loaded byte count
    if count != int64(len(state.assets)) {
        errstr := "Asset payload size (%v) differs from manifest tally (%v)."
        errstr += " Something is really, really wrong."
        errstr = fmt.Sprintf(errstr, len(state.assets), count)
        return debug(errors.New(errstr))
    }

    return nil
}

// Recursively verify an object.
func verifyObject(obj *Object) (count int64, err error) {
    // Directory?
    if obj.ModeBits.IsDir() {
        if obj.Size != 0 || obj.Offset != 0 || obj.Checksum != 0 {
            return 0, debug(errors.New("Directory object does not pass all sanity checks."))
        }
    } else {
        // File
        if obj.Size == 0 {
            if obj.Offset != 0 || obj.Checksum != 0 {
                return 0, debug(errors.New("File object does not pass all sanity checks."))
            }
        } else {
            if len(obj.Objects) != 0 {
                return 0, debug(errors.New("File object does not pass all sanity checks."))
            }
            data, err := getPayload(obj)
            if err != nil { return 0, err }
            h := crc32.NewIEEE()
            h.Write(data)
            if obj.Checksum != h.Sum32() {
                return 0, debug(errors.New("Checksum error."))
            }
            count += obj.Size
        }
    }

    // Verify child objects
    for i := 0; i < len(obj.Objects); i++ {
        bytes, err := verifyObject(&obj.Objects[i])
        if err != nil { return 0, debug(err) }
        count += bytes
    }

    return count, nil
}
