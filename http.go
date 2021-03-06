// http.go implements drop-in replacements for functions in the net/http
// package.

package caviar

import (
    "errors"
    "strings"
    "path"
    "path/filepath"
    "net/http"
)

// Replacement for net/http.Dir.
type Dir string

// Caviarized copy of net/http.Dir.Open()
func (d Dir) Open(name string) (http.File, error) {
    if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
        strings.Contains(name, "\x00") {
        return nil, debug(errors.New("http: invalid character in file path"))
    }
    dir := string(d)
    if dir == "" {
        dir = "."
    }
    f, err := Open(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name))))
    if err != nil {
        return nil, debug(err)
    }
    return f, nil
}
