package caviar

import (
    "path"
    "path/filepath"
    "strings"
)

var BinExtensions = [...]string {"bin", "exe", "elf"}
const CaviarExtension = "cvr"

func DetachedName(programpath string) string {
    program := filepath.Base(programpath)
    lprogram := strings.ToLower(program)
    for _, ext := range BinExtensions {
        if strings.HasSuffix(lprogram, "." + ext) {
            program = program[:len(program)-len(ext)]
        }
    }
    program += "." + CaviarExtension
    return path.Join(filepath.Dir(programpath), program)
}
