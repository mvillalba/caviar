// This little program creates a very simple (and very fast) Web server to
// serve all files in the current working directory.
package main

import (
    "log"
    "net/http"
    "github.com/mvillalba/caviar"
)

func main() {
    // caviar.Dir is a drop-in replacement for http.Dir
    err := http.ListenAndServe(":8000", http.FileServer(caviar.Dir(".")))
    if err != nil { log.Fatal(err) }
}
