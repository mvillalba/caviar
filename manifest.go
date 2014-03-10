package caviar

import (
    "fmt"
    "errors"
)

type Manifest struct {
    Magic       string
    Comment     string
    AssetRoot   DirectoryInfo
    Files       []FileInfo
}

type DirectoryInfo struct {
    Name        string
    Directories []DirectoryInfo
    Files       []int
}

type FileInfo struct {
    Name        string
    Size        uint
}

const MANIFEST_MAGIC = "CAVIAR"
const ASSETROOT_MAGIC = "ROOT"

func processManifest(manifest *Manifest, root *directoryInfo) (int, error) {
    // Verify magic
    if manifest.Magic != MANIFEST_MAGIC {
        errstr := "Container has invalid magic value (expected %v, got %v)."
        errstr = fmt.Sprintf(errstr, MANIFEST_MAGIC, manifest.Magic)
        return 0, errors.New(errstr)
    }

    // Verify asset root name tag magic
    if manifest.AssetRoot.Name != ASSETROOT_MAGIC {
        errstr := "Container has invalid magic value (expected %v, got %v)."
        errstr = fmt.Sprintf(errstr, ASSETROOT_MAGIC, manifest.AssetRoot.Name)
        return 0, errors.New(errstr)
    }

    // Process manifest
    count, err := processDirectory(manifest.Files, &manifest.AssetRoot, root)
    if err != nil { return 0, err }

    return count, nil
}

func processDirectory(manfiles []FileInfo, manroot *DirectoryInfo, root *directoryInfo) (int, error) {
    count := 0

    // General properties
    root.name = manroot.Name

    // Parent link (used to handle '.' and '..' path segments)
    if root.name == ASSETROOT_MAGIC {
        root.parent = nil
    }

    // Files
    for _, idx := range manroot.Files {
        newfile := new(fileInfo)
        newfile.parent = root
        newfile.name = manfiles[idx].Name
        newfile.size = manfiles[idx].Size
        newfile.offset = calcOffset(manfiles, idx)
        root.files = append(root.files, *newfile)
        count += int(newfile.size)
    }

    // Directories
    for _, entry := range manroot.Directories {
        lentry := new(directoryInfo)
        lentry.parent = root
        subcount, err := processDirectory(manfiles, &entry, lentry)
        if err != nil { return 0, err }
        root.directories = append(root.directories, *lentry)
        count += subcount
    }

    return count, nil
}

func calcOffset(manfiles []FileInfo, idx int) (count uint) {
    for i := 0; i < idx; i++ {
        count += manfiles[i].Size
    }
    return count
}
