Caviar
======

Caviar is a resource packer for Go. It will essentially pack a bunch of
resources/assets that you would normally deploy along with your executable into
a custom ZIP file which will then bundled with your program's executable (or
not, your choice).

*NOTE: UNDER DEVELOPMENT. NOT READY FOR PRIMETIME.*


Install
-------
Just run:

 $ go get github.com/mvillalba/caviar{,/caviarize,/cavundle}


Usage
-----
Converting your program to use Caviar is very straightforward, just import
"github.com/mvillalba/caviar", change any calls to os.Open/OpenFile to
caviar.Open/OpenFile, and run the bundled `cavundle` utility on your compiled
executables. That's it.

During runtime, Caviar will attempt to load bundled resources from the running
executable and failing that from a detached container (executable-name.cvr).

Your program will still run in the event a Caviar bundle can't be loaded.
Caviar will simply pass through your Open/OpenFile calls to the os package
transparently. This is very useful for development. And the same goes for
opening files not present in the bundle.

Got dependencies that need to read various files and won't take an io.Reader
(I'm looking at you, Revel framework)? You can use the bundled tool
`caviarize` which will attempt to automatically patch any package (and all its
dependencies) to load their files via Caviar's API like this:

 $ caviarize github.com/revel/revel

See the examples directory for a handful of working toy program examples.

*NOTE: In order to generate attached bundles (program = program + asset
bundle), cavundle needs the to execute `zip` program due to a shortcoming with
Go's ZIP library*

*NOTE: Caviar is designed with long-running processes (such as Web apps) that
need to have quick access to their assets/resources in mind and this has some
consequences. Namely, Caviar will load all assets to RAM on startup and it will
keep them there.*

*NOTE: This is an early version of Caviar and no cross-platform testing has been
done. It works on Linux (and probably other UNIX variants), but using it on
Windows will likely require some work (case-insensitive matches, possible
hard-coded paths, etc.). You are welcome to submit a patch.*

[![GoDoc](https://godoc.org/github.com/mvillalba/caviar?status.png)](https://godoc.org/github.com/mvillalba/caviar)


Contact
-------
Martín Raúl Villalba <martin@martinvillalba.com>
http://www.martinvillalba.com/


TODO
----
 * Support for unpacking to a TMP dir instead of RAM (for things like Gtk where
   we depend on C libraries that can't be patched with caviarize). Don't forget
   to implement a clean-up function to be run when the program exits!
 * Support for a custom, dead-simple binary format to replace ZIP files (Go's
   Zips don't support writing uncompressed data and TARs don't feel like a good
   replacement. The replacement format would be as follows:
   [EXECUTABLE][MAGIC-1][MANIFEST][ASSETS][MANIFEST-LEN][ASSETS-LEN][MAGIC-2]
   Reading it is a simple matter of:
    1. Read and verify MAGIC-2.
    2. Add ASSETS-LEN and MANIFEST-LEN and use them to calculate start offset
       of Caviar container (MAGIC-1 start).
    3. Read and verify MAGIC-1.
    4. Read both MANIFEST and ASSETS.
    5. Store ASSETS.
    6. Process MANIFEST.
   I may optionally consider adding a FLAGS byte somewhere to specify the
   compression algorithm used for MANIFEST and ASSETS. Options will probably be
   DEFLATE, GZIP, and NONE.
   No format versioning or backwards compatibility is needed as it's assumed
   the running version of Caviar will both be producing the asset bundle and
   reading it during runtime. The supported optimization profiles could be:
   FAST     MANIFEST (NONE), ASSETS (NONE)
   TINY     MANIFEST (GZIP), ASSETS (GZIP)
   NORM     MANIFEST (NONE), ASSETS (DEFLATE)
 * Make the manifest use Protocol Buffers instead of Gob.
 * Cross-platform support.
 * Some functions in the "os" file-related API return an os.PathError instead
   of an error type. Caviar should mimick this behavior.
 * Practical examples.
 * Tests, tests, tests.
 * Preserve metadata other than the file name. I'm thinking creation time,
   modification time, and permission bits.
 * Handle absolute paths better in caviarOpen(). At the moment, any absolute
   path must match the path string returned by osext.ExecutableFolder() exactly
   to be considered to be inside the asset root path, so something like
   '/home/martin/wks/../go/myprogram/data.bin' will be passed to os.Open (and
   error out with a file not found error) while
   '/home/martin/wks/go/myprogram/data.bin' will succeed.
 * The code needs a general clean-up, but fo it after mostly everything else on
   this list has been implemented.
 * Can't be bothered now, but there is a hack (find “BEGIN HACK”) in init.go
   to patch a bug somewhere (probably in processManifest() or cavundle) that
   creates a sort of pseuso-root dir in state.assets that's kinda broken.
 * Invoking cavundle with multiple paths that contain files and/or directories
   with repeated names (i.e revel/README, martini/README, etc.) and
   cherrypicking disabled is likely broken. Fix it.
 * Implement a Caviar FS walk function to shadow path/filepath.Walk().
 * There is a bit of a type casting mess. Make everything use int64 and be done
   with it.
 * Port caviarize to Go.
 * Get auxiliary os methods working (Stat, Lstat, etc.).
 * Make Revel work.
 * Make Martini work.
 * Allow directories to be open (Martini seems to need this).
 * Make sure all functions only operate when Caviar is ready and error out
   otherwise.
 * Setup Travis CI/Wercker and Godoc.
 * Make sure no CaviarFile related function allows itself to be called on a
   closed file. Should return os.ErrInvalid.
 * Apparently, in for…range…{} constructs, range makes a copy of the object
   being looped, which is really bad for the object tree which could
   potentially be pretty big. Go through the code and make the loops not use
   range when iterating over the object tree. Or perhaps this is pointless, as
   the objects themselves hold children on a slice which is a reference type,
   right?
 * Functions such as CaviarFile.Readdir() should merge their own output with
   that of the native OS and merge them together.
 * Run some benchmarks on just how much faster (or slower) Caviar is relative
   to the native OS both for files in and out of the kernel's disk cache.
 * More documentation.
 * Move issues to GitHub's bug tracker.
