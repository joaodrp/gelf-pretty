package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

var (
	// Flags
	versionFlag = flag.Bool("version", false, "Show version information")

	// Version is the binary SemVer version (latest git tag)
	Version string
	// BuildCommit is the hash of the git commit used to build the binary
	BuildCommit string
	// BuildTime is the binary build timestamp
	BuildTime string

	// Map syslog levels to a human readable name
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
)

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

// versionInfo will display the compile/build-time version variables. This is
// available through the `version` flag:
//
// $ gelf-pretty -version
//
//           Version:  0.1.0
// Build Commit Hash:  640197df9b907efe9bfdf8ac2914b28a3ec9b8ef
//        Build Time:  2019-03-30T12:48:27Z
//
func versionInfo() *bytes.Buffer {
	b := new(bytes.Buffer)
	w := new(tabwriter.Writer)
	w.Init(b, 0, 0, 0, ' ', tabwriter.AlignRight)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Version:", "\t", Version)
	_, _ = fmt.Fprintln(w, "Build Commit Hash:", "\t", BuildCommit)
	_, _ = fmt.Fprintln(w, "Build Time:", "\t", BuildTime)
	_, _ = fmt.Fprintln(w)
	_ = w.Flush()
	return b
}

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Print(versionInfo())
		os.Exit(0)
	}

	pp := newPrettyPrinter(os.Stdin, os.Stdout)
	if err := pp.run(); err != nil {
		panic(err)
	}
}
