package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "archive/zip"
    "flag"
    "log"
    "bytes"
    "encoding/gob"
    "caviar"
    "io/ioutil"
    "strings"
    "errors"
)

const MANIFEST_COMMENT =
`Generated by the Caviar resource packer for Go (http://github.com/mvillalba/caviar).
Copyright © 2014 Martín Raúl Villalba (http://www.martinvillalba.com/).`

type Args struct {
    cherrypick  bool
    detached    bool
    executable  string
    paths       []string
}

func parseArgs() (a Args) {
    cphelp := "add asset paths as sub-directories (rather than merge all contained files and directories across asset paths under one directory)."
    dthelp := "produce detached asset container."
    flag.BoolVar(&a.cherrypick, "cherrypick", false, cphelp)
    flag.BoolVar(&a.detached, "detached", false, dthelp)
    flag.Parse()

    if len(flag.Args()) < 2 {
        fmt.Println("Cavundle is part of the Caviar resource packer for Go (http://github.com/mvillalba/caviar).")
        fmt.Println("Copyright © 2014 Martín Raúl Villalba <martin@martinvillalba.com>")
        fmt.Println("")
        fmt.Println("Usage: %s [--cherrypick] EXECUTABLE ASSET-PATH-1[...ASSET-PATH-N]")
        flag.PrintDefaults()
        os.Exit(1)
    }

    a.executable = flag.Args()[0]
    for _, path := range flag.Args()[1:] {
        path, err := filepath.Abs(path)
        if err != nil { log.Fatal(err) }
        a.paths = append(a.paths, path)
    }

    return a
}


func getDir(path, prefix string, manifest *caviar.Manifest) *caviar.DirectoryInfo {
    if !strings.HasPrefix(path, prefix) { log.Fatal("Internal error.") }
    curdir := &manifest.AssetRoot
    for _, segment := range strings.Split(path[len(prefix):], "/") {
        // Find segment in curdir
        found := false
        var newcurdir *caviar.DirectoryInfo
        for i, entry := range curdir.Directories {
            if entry.Name != segment { continue }
            found = true; newcurdir = &curdir.Directories[i]; break
        }

        // Add dir representing segment to curdir
        if !found {
            var newdir caviar.DirectoryInfo
            newdir.Name = segment
            curdir.Directories = append(curdir.Directories, newdir)
            newcurdir = &curdir.Directories[len(curdir.Directories)-1]
        }

        // Set curdir to found/added directory entry
        curdir = newcurdir
    }
    return curdir
}

func processAssets(args Args) (caviar.Manifest, []byte, error) {
    // Init
    buf := new(bytes.Buffer)
    var manifest caviar.Manifest
    manifest.Magic = caviar.MANIFEST_MAGIC
    manifest.Comment = MANIFEST_COMMENT
    manifest.AssetRoot.Name = caviar.ASSETROOT_MAGIC

    // Walk
    for _, path := range args.paths {
        dir := path
        if args.cherrypick { dir = filepath.Dir(path) }

        walkfunc := func(path string, info os.FileInfo, err error) error {
            if err != nil { log.Fatal(err) }
            if info.IsDir() {
                // Add dir to manifest
                _ = getDir(path, dir, &manifest)
            } else {
                // File info
                assetfile := caviar.FileInfo{info.Name(), uint(info.Size())}
                manifest.Files = append(manifest.Files, assetfile)
                assetindex := len(manifest.Files) - 1

                // Asset data
                data, err := ioutil.ReadFile(path)
                if err != nil { log.Fatal(err) }
                buf.Write(data)

                // Directory info
                assetdir := getDir(filepath.Dir(path), dir, &manifest)
                assetdir.Files = append(assetdir.Files, assetindex)
            }
            return nil
        }

        filepath.Walk(path, walkfunc)
    }

    return manifest, buf.Bytes(), nil
}

func main() {
    args := parseArgs()

    // Container
    buf := new(bytes.Buffer)
    zw := zip.NewWriter(buf)

    // Pack assets
    manifest, assets, err := processAssets(args)
    if err != nil { log.Fatal(err) }

    f, _ := zw.Create("Manifest.gob")
    enc := gob.NewEncoder(f)
    err = enc.Encode(manifest)
    if err != nil { log.Fatal(err) }

    f, _ = zw.Create("Assets.bin")
    _, err = f.Write(assets)
    if err != nil { log.Fatal(err) }

    // Clean up
    err = zw.Close()
    if err != nil {
        log.Fatal(err)
    }

    // Dump buffer
    path, err := filepath.Abs(args.executable)
    if err != nil { log.Fatal(err) }

    var fp *os.File
    if args.detached {
        path = caviar.DetachedName(path)
        fp, err = os.OpenFile(path, os.O_WRONLY | os.O_CREATE, 0664)
    } else {
        fp, err = os.OpenFile(path, os.O_WRONLY | os.O_APPEND, 0664)
    }
    if err != nil { log.Fatal(err) }

    buf.WriteTo(fp)

    fp.Sync()
    fp.Close()

    // Re-align container
    errprefix := "Zip align error: "
    cmd := exec.Command("zip", "-A", path)
    err = cmd.Start()
    if err != nil { log.Fatal(errors.New(errprefix + err.Error())) }
    err = cmd.Wait()
    if err != nil { log.Fatal(errors.New(errprefix + err.Error())) }
}
