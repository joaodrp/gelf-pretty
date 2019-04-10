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

	// Additional fields that have a special behaviour
	specialFields = map[string]string{
		"app":    "_app",
		"logger": "_logger",
	}
)

const timeFormat = "2006-01-02 15:04:05.000"

type fullMessage string

func (fm fullMessage) String() string {
	if len(fm) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\t")
	sb.WriteString(strings.ReplaceAll(string(fm), "\n", "\n\t"))

	return sb.String()
}

type unixTimestamp float64

func (t unixTimestamp) String() string {
	sec, dec := math.Modf(float64(t))
	tm := time.Unix(int64(sec), int64(dec*(1e9)))
	str := tm.Format(timeFormat)
	return fmt.Sprintf("[%s]", str)
}

type syslogLevel int

func (l syslogLevel) String() string {
	return levelToName[int(l)]
}

type additionalField struct {
	key   string
	value interface{}
}

func (af additionalField) String() string {
	return fmt.Sprintf("%s=%v", strings.TrimPrefix(af.key, "_"), af.value)
}

func (af additionalField) special() bool {
	for _, v := range specialFields {
		if af.key == v {
			return true
		}
	}
	return false
}

type additionalFields []additionalField

func (afs additionalFields) String() string {
	if len(afs) == 0 {
		return ""
	}
	var sb strings.Builder
	i := 0
	for _, af := range afs {
		if af.special() {
			continue
		}
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(af.String())
		i++
	}
	return sb.String()
}

type gelf struct {
	version          string
	host             string
	shortMessage     string
	fullMessage      fullMessage
	timestamp        unixTimestamp
	level            syslogLevel
	additionalFields additionalFields
}

type dict map[string]interface{}

func (d dict) findByKeyAndCastToString(key string, required bool) (string, error) {
	val, ok := d[key]
	if !ok && required {
		return "", fmt.Errorf("%s not found", key)
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s is not a valid string", key)
	}
	delete(d, key)
	return s, nil
}

func (d dict) findByKeyAndCastToFloat64(key string, required bool) (float64, error) {
	val, ok := d[key]
	if !ok && required {
		return 0, fmt.Errorf("%s not found", key)
	}
	n, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf("%s is not a valid number", key)
	}
	delete(d, key)
	return n, nil
}

func (g *gelf) UnmarshalJSON(data []byte) error {
	d := dict{}
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	v, err := d.findByKeyAndCastToString("version", true)
	if err != nil {
		return err
	}
	g.version = v

	h, err := d.findByKeyAndCastToString("host", true)
	if err != nil {
		return err
	}
	g.host = h

	sm, err := d.findByKeyAndCastToString("short_message", true)
	if err != nil {
		return err
	}
	g.shortMessage = sm

	fm, err := d.findByKeyAndCastToString("full_message", false)
	if err != nil {
		return err
	}
	g.fullMessage = fullMessage(fm)

	t, err := d.findByKeyAndCastToFloat64("timestamp", true)
	if err != nil {
		return err
	}
	g.timestamp = unixTimestamp(t)

	l, err := d.findByKeyAndCastToFloat64("level", true)
	if err != nil {
		return err
	}
	g.level = syslogLevel(l)

	for k, v := range d {
		if strings.HasPrefix(k, "_") {
			af := additionalField{key: k, value: v}
			g.additionalFields = append(g.additionalFields, af)
		}
	}
	return nil
}

func (g *gelf) findAdditionalFieldValueByKey(key string) string {
	for _, af := range g.additionalFields {
		if af.key == key {
			return af.value.(string)
		}
	}
	return ""
}

func (g *gelf) app() string {
	return g.findAdditionalFieldValueByKey(specialFields["app"])
}

func (g *gelf) logger() string {
	return g.findAdditionalFieldValueByKey(specialFields["logger"])
}

func (g *gelf) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s: ", g.timestamp, g.level))

	a := g.app()
	if len(a) > 0 {
		sb.WriteString(a)
	}

	l := g.logger()
	if len(l) > 0 {
		if len(a) > 0 {
			sb.WriteString("/")
		}
		sb.WriteString(l)
	}

	if h := g.host; len(h) > 0 {
		if len(a) > 0 || len(l) > 0 {
			sb.WriteString(" on ")
		}
		sb.WriteString(h)
	}

	sb.WriteString(fmt.Sprintf(": %s", g.shortMessage))
	sb.WriteString(fmt.Sprintf(" %s", g.additionalFields))
	sb.WriteString(g.fullMessage.String())

	return sb.String()
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

func (h *prettyPrinter) processLine(b []byte) error {
	g := &gelf{}
	if err := json.Unmarshal(b, g); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(h.writer, g); err != nil {
		return err
	}
	return nil
}

func (h *prettyPrinter) run() error {
	for h.reader.Scan() {
		b := h.reader.Bytes()
		if len(b) == 0 {
			continue
		}
		if err := h.processLine(b); err != nil {
			if _, err := fmt.Fprintln(h.writer, string(b)); err != nil {
				return err
			}
			continue
		}
	}
	if err := h.reader.Err(); err != nil {
		return err
	}
	return nil
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
