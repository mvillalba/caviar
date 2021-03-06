// This example showcases a trivial Martini app mostly serving static files
// from a bundle. Martini needs to be patched to be able to serve it's files
// through Caviar's API which we'll do automatically with the caviarize
// utility.
package main

import (
    "github.com/codegangsta/martini"
)

func main() {
    m := martini.Classic()
    m.Get("/", func() string {
        return "Hello, Caviarized Martini!"
    })
    m.Run()
}

