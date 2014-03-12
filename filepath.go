// filepath.go implements drop-in replacement functions for the path/filepath
// package.

package caviar

import (
    "errors"
    "path/filepath"
)

func Walk(root string, walkFn filepath.WalkFunc) error {
    return errors.New("Walk(): Not Implemented.") // TODO
}
