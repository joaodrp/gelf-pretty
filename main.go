package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// Flags
var versionFlag = flag.Bool("v", false, "Show version information")

// Versioning variables
var (
	// Version is the SemVer application version (set to the latest git tag)
	Version string
	// BuildCommit is the hash of the git commit on which the binary was built
	BuildCommit string
	// BuildBranch is the branch from which the binary was built
	BuildBranch string
	// BuildTime is the build timestamp
	BuildTime string
	// BuildAuthor is the git email address of the user who built the binary
	BuildAuthor string
)

// printVersion will display the compile/build-time versioning variables. This
// is available through the `version` flag:
//
// $ gelf-pretty -v
//
//           Version:  0.1.0
// Build Commit Hash:  640197df9b907efe9bfdf8ac2914b28a3ec9b8ef
//      Build Branch:  master
//        Build Time:  2019-03-30T12:48:27Z
//      Build Author:  joaodrp@gmail.com
//
func printVersion() {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Version:", "\t", Version)
	fmt.Fprintln(w, "Build Commit Hash:", "\t", BuildCommit)
	fmt.Fprintln(w, "Build Branch:", "\t", BuildBranch)
	fmt.Fprintln(w, "Build Time:", "\t", BuildTime)
	fmt.Fprintln(w, "Build Author:", "\t", BuildAuthor)
	fmt.Fprintln(w)
	w.Flush()
}

func main() {
	flag.Parse()

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}
}
