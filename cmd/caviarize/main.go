// Caviarize is meant to help patch dependencies that need to access files and
// directories you'd like to put in a Caviar bundle. It's recommended you run
// it on a clean $GOPATH, like so:
//
//      GOPATH=/tmp/throwawaygopath caviarize github.com/codegangsta/martini
//
// Caviarize will then fetch Martini and patch it (along with all its
// dependencies.
//
// I recommend you don't use caviarize on your own code, but rather make it
// Caviar-compatible manually. Caviarize is a bit of a blunt tool; it doesn't
// actually parse Go code, it just does regex-based search and replace and
// produces rather unpleasant looking code which is likely not what you want.
package main

import (
    "fmt"
    "os"
    "os/exec"
    "flag"
    "errors"
    "regexp"
    "path"
    "path/filepath"
    "crypto/md5"
    "io"
    "io/ioutil"
    "bufio"
    "strings"
)

type Args struct {
    nodep       bool
    importpath  string
}

func parseArgs() (a Args) {
    ndhelp := "don't patch dependencies."
    flag.BoolVar(&a.nodep, "nodep", false, ndhelp)
    flag.Parse()

    if len(flag.Args()) != 1 {
        fmt.Println("Caviarize is part of the Caviar resource packer for Go (http://github.com/mvillalba/caviar).")
        fmt.Println("Copyright © 2014 Martín Raúl Villalba <martin@martinvillalba.com>")
        fmt.Println("")
        fmt.Printf("Usage: %s [OPTIONS] IMPORT-PATH\n", os.Args[0])
        flag.PrintDefaults()
        os.Exit(1)
    }

    a.importpath = flag.Arg(0)

    return a
}

var patchList = [...]string{
    "os.Open",
    "os.OpenFile",
    "os.Lstat",
    "os.Stat",
    "filepath.Walk",
    "filepath.Glob",
    "http.Dir",
    "ioutil.ReadDir",
    "ioutil.ReadFile",
}

var patchFixups = [...]string{
    "os.FileInfo",
}

var pkgList = [...]string{
    "os.Open",
    "path/filepath.Walk",
    "net/http.Redirect",
    "io/ioutil.ReadFile",
}

var importregex = regexp.MustCompile(`^[\t ]*package[\t ]+([a-zA-Z0-9]+)[\t ]*$`)
const caviarImportPath = "github.com/mvillalba/caviar"

func patchFile(fpath string, info os.FileInfo, err error) error {
    // Pass along errors
    if err != nil { return err }

    // We only want to patch Go source files
    if info.IsDir() { return nil }

    // Go code?
    if filepath.Ext(fpath) != ".go" { return nil }

    // Yes, Go code! Load it!
    bytes, err := ioutil.ReadFile(fpath)
    if err != nil { return err }
    code := string(bytes)

    // Scan code for function calls that need replacing
    patchit := false
    for _, match := range patchList {
        if strings.Index(code, match) != -1 {
                if strings.Index(code, caviarImportPath) == -1 {
                patchit = true
                break
            }
        }
    }

    // Log
    if !patchit {
        fmt.Println("    SKIP", fpath)
        return nil
    }
    fmt.Println("    PATCH", fpath)

    // Replace function calls
    for _, fixup := range patchFixups {
        code = strings.Replace(code, fixup, getMd5(fixup), -1)
    }

    for _, find := range patchList {
        tmp := strings.Split(find, ".")
        replace := "caviar." + tmp[1]
        code = strings.Replace(code, find, replace, -1)
    }

    for _, fixup := range patchFixups {
        code = strings.Replace(code, getMd5(fixup), fixup, -1)
    }

    // Inject Caviar import
    code = strings.Replace(code, "\r\n", "\n", -1)
    code = strings.Replace(code, "\r", "\n", -1)
    newcode := ""
    for _, line := range strings.Split(code, "\n") {
        line = importregex.ReplaceAllString(line, "package $1\nimport \"" + caviarImportPath + "\"")
        newcode += line + "\n"
    }
    code = newcode

    // Inject “imported but not used” compile-time error workaround
    for _, pkg := range pkgList {
        t := strings.Split(pkg, ".")
        pkgimp := t[0]
        pkgfunc := t[1]
        t = strings.Split(pkgimp, "/")
        pkgref := t[len(t)-1]
        match := "\"" + pkgimp + "\""
        if strings.Index(code, match) != -1 {
            code += "\nvar _ = " + pkgref + "." + pkgfunc
        }
    }

    // Save patched source file
    err = ioutil.WriteFile(fpath, []byte(code), info.Mode())
    if err != nil { return err }

    return nil
}

func getMd5(s string) string {
    h := md5.New()
    io.WriteString(h, s)
    return fmt.Sprintf("%x", h.Sum(nil))
}

func patchPackage(pkg string) error {
    // Log
    fmt.Println("PKG", pkg)

    // Build package path
    pkgpath := os.Getenv("GOPATH")
    if pkgpath == "" {
        return errors.New("Could not resolve $GOPATH env var.")
    }
    pkgrep := strings.Replace(pkg, "/", string(os.PathSeparator), -1)
    pkgpath = path.Join(pkgpath, "src", pkgrep)
    pkgpath, err := filepath.Abs(pkgpath)
    if err != nil { return err }
    pkgpath = path.Clean(pkgpath)

    // Patch'em, cowboy! Patch'em!
    err = filepath.Walk(pkgpath, patchFile)
    if err != nil { return err }

    return nil
}

func main() {
    args := parseArgs()

    // Install Caviar itself, if not already present
    fmt.Println("GET", caviarImportPath)
    cmd := exec.Command("go", "get", "-v", "-d", caviarImportPath)
    err := cmd.Run()
    if err != nil { die(err) }

    // Download package and associated dependencies.
    var pkglist []string

    cmd = exec.Command("go", "get", "-v", "-d", args.importpath)
    stderr, err := cmd.StderrPipe()
    if err != nil { die(err) }
    err = cmd.Start()
    if err != nil { die(err) }

    buf := bufio.NewReader(stderr)
    for true {
        line, err := buf.ReadString('\n')
        if err != nil && err != io.EOF { die(err) }
        parts := strings.Split(line, " ")
        if parts[0] == "" { break }
        pkglist = append(pkglist, parts[0])
        fmt.Println("GET", parts[0])
    }

    err = cmd.Wait()
    if err != nil { die(err) }

    // Patch all relevant Go source files for all downloaded packages.
    for _, pkg := range pkglist {
        err = patchPackage(pkg)
        if err != nil { die(err) }
    }

    // Build and install Caviarized package(s)
    fmt.Println("INSTALL", args.importpath)
    cmd = exec.Command("go", "install", "-v", args.importpath)
    err = cmd.Run()
    if err != nil { die(err) }
}

func die(err error) {
    fmt.Println("ERROR", err)
    os.Exit(1)
}
