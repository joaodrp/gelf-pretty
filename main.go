package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"text/tabwriter"
	"time"
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

const (
	timeFormat = "2006-01-02 15:04:05.000"
	appKey     = "_app"
	loggerKey  = "_logger"
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
		b := h.reader.Bytes()
		if len(b) == 0 {
			continue
		}
		if err := h.processLine(b); err != nil {
			return err
		}
	}
	if err := h.reader.Err(); err != nil {
		return err
	}
	return nil
}

func msToTime(ms float64) time.Time {
	sec, dec := math.Modf(ms)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}

func (h *prettyPrinter) processLine(b []byte) error {
	g := &gelf{}
	if err := json.Unmarshal(b, g); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(h.writer, g.pretty()); err != nil {
		return err
	}
	return nil
}

type gelf struct {
	version          string
	host             string
	shortMessage     string
	fullMessage      string
	timestamp        float64
	Level            int
	additionalFields []additionalField
}

type additionalField struct {
	key   string
	value interface{}
}

func findByKeyAndCastToString(m map[string]interface{}, key string, required bool) (string, error) {
	val, ok := m[key]
	if !ok && required {
		return "", fmt.Errorf("%s not found", key)
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s is not a valid string", key)
	}
	return s, nil
}

func findByKeyAndCastToFLoat64(m map[string]interface{}, key string, required bool) (float64, error) {
	val, ok := m[key]
	if !ok && required {
		return 0, fmt.Errorf("%s not found", key)
	}
	n, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf("%s is not a valid number", key)
	}
	return n, nil
}

func (g *gelf) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	v, err := findByKeyAndCastToString(m, "version", true)
	if err != nil {
		return err
	}
	g.version = v
	delete(m, "version")

	h, err := findByKeyAndCastToString(m, "host", true)
	if err != nil {
		return err
	}
	g.host = h
	delete(m, "host")

	sm, err := findByKeyAndCastToString(m, "short_message", true)
	if err != nil {
		return err
	}
	g.shortMessage = sm
	delete(m, "short_message")

	fm, err := findByKeyAndCastToString(m, "full_message", false)
	if err != nil {
		return err
	}
	g.fullMessage = fm
	delete(m, "full_message")

	t, err := findByKeyAndCastToFLoat64(m, "timestamp", true)
	if err != nil {
		return err
	}
	g.timestamp = t
	delete(m, "timestamp")

	l, err := findByKeyAndCastToFLoat64(m, "level", true)
	if err != nil {
		return err
	}
	g.Level = int(l)
	delete(m, "level")

	for k, v := range m {
		if strings.HasPrefix(k, "_") {
			af := additionalField{key: strings.TrimPrefix(k, "_"), value: v}
			g.additionalFields = append(g.additionalFields, af)
		}
	}
	return nil
}

// Implement the String interface for pretty printing Items
// func (i Item) String() string {
// 	return fmt.Sprintf("Item: %s, %d", i.Item1, i.Item2)
// }

func (g *gelf) prettyTime() string {
	t := msToTime(g.timestamp).Format(timeFormat)
	return fmt.Sprintf("[%s]", t)
}

func (g *gelf) prettyLevel() string {
	return levelToName[g.Level]
}

func (g *gelf) findAddFieldValueByKey(key string) interface{} {
	for _, af := range g.additionalFields {
		if af.key == key {
			return af.value
		}
	}
	return ""
}

func (g *gelf) prettyApp() string {
	return fmt.Sprintf("%v", g.findAddFieldValueByKey(appKey))
}

func (g *gelf) prettyLogger() string {
	return fmt.Sprintf("%v", g.findAddFieldValueByKey(loggerKey))
}

func (g *gelf) prettyHost() string {
	return g.host
}

func (g *gelf) prettyShortMessage() string {
	return g.shortMessage
}

func (g *gelf) prettyFullMessage() string {
	if len(g.fullMessage) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\t")
	sb.WriteString(strings.ReplaceAll(g.fullMessage, "\n", "\n\t"))

	return sb.String()
}

func (g *gelf) prettyAdditionalFields() string {
	if len(g.additionalFields) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, af := range g.additionalFields {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%s=%v", af.key, af.value))
	}
	return sb.String()
}

func (g *gelf) pretty() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s: ", g.prettyTime(), g.prettyLevel()))

	a := g.prettyApp()
	if len(a) > 0 {
		sb.WriteString(a)
	}

	l := g.prettyLogger()
	if len(l) > 0 {
		if len(a) > 0 {
			sb.WriteString("/")
		}
		sb.WriteString(l)
	}

	if h := g.prettyHost(); len(h) > 0 {
		if len(a) > 0 || len(l) > 0 {
			sb.WriteString("on ")
		}
		sb.WriteString(h)
	}

	sb.WriteString(fmt.Sprintf(": %s", g.prettyShortMessage()))
	if af := g.prettyAdditionalFields(); len(af) > 0 {
		sb.WriteString(fmt.Sprintf(" %s", af))
	}

	sb.WriteString(g.prettyFullMessage())

	return sb.String()
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
