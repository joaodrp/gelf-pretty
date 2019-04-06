package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// Flags
var versionFlag = flag.Bool("v", false, "Show version information")

// Version variables
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

var (
	// Map syslog levels to human readable names
	levelToName map[int]string
)

func init() {
	levelToName = map[int]string{
		0: "EMERGENCY",
		1: "ALERT",
		2: "CRITICAL",
		3: "ERROR",
		4: "WARNING",
		5: "NOTICE",
		6: "INFO",
		7: "DEBUG",
	}
}

type prettyPrinter struct {
	reader *bufio.Scanner
	writer io.Writer
}

func newPrettyPrinter(r io.Reader, w io.Writer) *prettyPrinter {
	return &prettyPrinter{
		reader: bufio.NewScanner(r),
		writer: w,
	}
}

func (h *prettyPrinter) run() error {
	for h.reader.Scan() {
		t := h.reader.Text()
		if len(t) == 0 {
			continue
		}
		if err := h.processLine(t); err != nil {
			return err
		}
	}
	if err := h.reader.Err(); err != nil {
		return err
	}
	return nil
}

func (h *prettyPrinter) processLine(l string) error {
	_, err := fmt.Fprintln(h.writer, l)
	return err
}

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

func init() {
	levelToName = map[int]string{
		0: "EMERGENCY",
		1: "ALERT",
		2: "CRITICAL",
		3: "ERROR",
		4: "WARNING",
		5: "NOTICE",
		6: "INFO",
		7: "DEBUG",
	}
}

func main() {
	flag.Parse()
	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	pp := newPrettyPrinter(os.Stdin, os.Stdout)
	if err := pp.run(); err != nil {
		panic(err)
	}
}
