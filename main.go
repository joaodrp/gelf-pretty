package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

var (
	// versionFlag is a flag to display version information.
	versionFlag = flag.Bool("version", false, "Show version information")
	// noColorFlag is a flag to disable colored output.
	noColorFlag = flag.Bool("no-color", false, "Disable color output")

	// version is the binary SemVer version (latest git tag).
	version string
	// commit is the hash of the git commit used to build the binary.
	commit string
	// date is the binary build timestamp.
	date string

	// levelToName maps syslog levels to a human readable name.
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
	// levelToColor maps syslog levels to colors.
	levelToColor = map[int]color.Attribute{
		0: color.FgHiRed,
		1: color.FgHiRed,
		2: color.FgHiRed,
		3: color.FgRed,
		4: color.FgYellow,
		5: color.FgYellow,
		6: color.FgGreen,
		7: color.FgCyan,
	}

	// reservedFields are GELF additional fields that have a special behaviour.
	reservedFields = map[string]string{
		"app":    "_app",    // identifies the application name
		"logger": "_logger", // identifies the application module or logger instance
	}
)

// timeFormat is the layout used to format the GELF `timestamp` field.
const timeFormat = "2006-01-02 15:04:05.000"

// shortMessage represents the GELF `short_message` field.
type shortMessage string

// String returns the prettified representation of the log short message.
func (msg shortMessage) String() string {
	return color.New(color.Bold).Sprint(string(msg))
}

// fullMessage represents the GELF `full_message` field.
type fullMessage string

// String returns the prettified representation of the log full message.
func (fm fullMessage) String() string {
	if len(fm) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\t")
	sb.WriteString(strings.ReplaceAll(string(fm), "\n", "\n\t"))

	return sb.String()
}

// timestamp represents the GELF `timestamp` field.
type timestamp struct {
	epoch    float64
	location *time.Location
}

// String returns the prettified representation of the log timestamp.
func (t timestamp) String() string {
	sec, dec := math.Modf(t.epoch)
	aux := time.Unix(int64(sec), int64(dec*(1e9))).In(t.location)
	return fmt.Sprintf("[%s]", aux.Format(timeFormat))
}

// syslogLevel represents the GELF `level` field.
type syslogLevel int

// String returns the prettified representation of the log level.
func (l syslogLevel) String() string {
	i := int(l)
	c := color.New(levelToColor[i], color.Bold)
	return c.Sprint(levelToName[i])
}

// additionalField represents a GELF `_*` additional field.
type additionalField struct {
	key   string
	value interface{}
}

// String returns the prettified representation of the additional field.
func (af additionalField) String() string {
	key := strings.TrimPrefix(af.key, "_")
	return fmt.Sprintf("%s=%v", color.MagentaString(key), af.value)
}

// reserved finds out if the additional field is reserved or not.
func (af additionalField) reserved() bool {
	for _, v := range reservedFields {
		if af.key == v {
			return true
		}
	}
	return false
}

// additionalFields represents a list of GELF `_*` additional fields.
type additionalFields []additionalField

// String returns the prettified representation of a list of additional fields.
func (afs additionalFields) String() string {
	if len(afs) == 0 {
		return ""
	}

	// sort fields by key for predictability/reproducibility
	sort.Slice(afs, func(i, j int) bool {
		return afs[i].key < afs[j].key
	})

	var sb strings.Builder
	i := 0
	for _, af := range afs {
		if af.reserved() {
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

// gelf represents a GELF log message.
type gelf struct {
	version          string
	host             string
	shortMessage     shortMessage
	fullMessage      fullMessage
	timestamp        timestamp
	level            syslogLevel
	additionalFields additionalFields
}

// dict is an utility alias for `map[string]interface{}`.
type dict map[string]interface{}

// _findByKey searches a dictionary by key. If found, the value is returned. An
// error is returned if the key was not found and it was said to be required.
func (d dict) _findByKey(key string, required bool) (interface{}, error) {
	val, ok := d[key]
	if !ok && required {
		return nil, fmt.Errorf("%s not found", key)
	}
	return val, nil
}

// findByKeyAndCastToString searches a dictionary by key and attempts to cast
// its value to string.
func (d dict) findByKeyAndCastToString(key string, required bool) (string, error) {
	val, err := d._findByKey(key, required)
	if err != nil || val == nil {
		return "", err
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s is not a valid string", key)
	}
	delete(d, key)
	return s, nil
}

// findByKeyAndCastToFloat64 searches a dictionary by key and attempts to cast
// its value to float64.
func (d dict) findByKeyAndCastToFloat64(key string, required bool) (float64, error) {
	val, err := d._findByKey(key, required)
	if err != nil || val == nil {
		return 0, err
	}
	n, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf("%s is not a valid number", key)
	}
	delete(d, key)
	return n, nil
}

// UnmarshalJSON unmarshal a JSON string to a gelf struct
func (g *gelf) UnmarshalJSON(data []byte) error {
	d := dict{}
	_ = json.Unmarshal(data, &d) // if it gets here it never fails

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
	g.shortMessage = shortMessage(sm)

	fm, err := d.findByKeyAndCastToString("full_message", false)
	if err != nil {
		return err
	}
	g.fullMessage = fullMessage(fm)

	t, err := d.findByKeyAndCastToFloat64("timestamp", true)
	if err != nil {
		return err
	}
	g.timestamp.epoch = t

	l, err := d.findByKeyAndCastToFloat64("level", false)
	if err != nil {
		return err
	}
	if l == 0 {
		l = 1
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

// findAdditionalFieldValueByKey searches for an additional field by key and
// returns its value as a string.
func (g *gelf) findAdditionalFieldValueByKey(key string) string {
	for _, af := range g.additionalFields {
		if af.key == key {
			return af.value.(string)
		}
	}
	return ""
}

// app returns the reserved `_app` additional field value.
func (g *gelf) app() string {
	return g.findAdditionalFieldValueByKey(reservedFields["app"])
}

// logger returns the reserved `_logger` additional field value.
func (g *gelf) logger() string {
	return g.findAdditionalFieldValueByKey(reservedFields["logger"])
}

// String returns the prettified representation of a GELF message.
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

// prettyPrinter represents a pretty-printer for GELF messages.
type prettyPrinter struct {
	reader   *bufio.Scanner
	writer   io.Writer
	location *time.Location // used to format/localize timestamps
}

// newPrettyPrinter returns a new prettyPrinter.
func newPrettyPrinter(r io.Reader, w io.Writer, l *time.Location) *prettyPrinter {
	pp := prettyPrinter{
		reader: bufio.NewScanner(r),
		writer: w,
	}
	if l == nil {
		l = time.Local
	}
	pp.location = l
	return &pp
}

// processLine pretty-prints a GELF message line.
func (h *prettyPrinter) processLine(b []byte) error {
	g := &gelf{}
	if err := json.Unmarshal(b, g); err != nil {
		return err
	}
	g.timestamp.location = h.location
	if _, err := fmt.Fprintln(h.writer, g); err != nil {
		return err
	}
	return nil
}

// run starts the pretty-printer.
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

// versionInfo builds the compile/build-time version information. This is
// available through the `version` flag.
func versionInfo(w io.Writer) error {
	tw := new(tabwriter.Writer)
	tw.Init(w, 0, 0, 0, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "Version:", "\t", version)
	fmt.Fprintln(tw, "Build Commit Hash:", "\t", commit)
	fmt.Fprintln(tw, "Build Time:", "\t", date)
	fmt.Fprintln(tw)
	return tw.Flush()
}

// run will parse flags, build and start the prettifiedPrinter.
func run(r io.Reader, w io.Writer) error {
	flag.Parse()
	if *versionFlag {
		return versionInfo(w)
	}

	color.NoColor = *noColorFlag

	pp := newPrettyPrinter(r, w, nil)
	if err := pp.run(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		panic(err)
	}
}
