// filepath.go implements drop-in replacement functions for the path/filepath
// package.

package caviar

import (
    "errors"
    "path/filepath"
)

func Walk(root string, walkFn filepath.WalkFunc) error {
    return debug(errors.New("Walk(): Not Implemented.")) // TODO
}

func Glob(pattern string) (matches []string, err error) {
    return nil, debug(errors.New("Glob(): Not Implemented.")) // TODO
}
