// ioutil.go implements drop-in replacements for io/ioutil functions.

package caviar

import (
    "os"
    "errors"
)

// TODO
func ReadFile(filename string) ([]byte, error) {
    return nil, debug(errors.New("ReadFile(): Not Implemented."))
}

// TODO
func ReadDir(dirname string) ([]os.FileInfo, error) {
    return nil, debug(errors.New("ReadFile(): Not Implemented."))
}
