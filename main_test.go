package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fatih/color"

	"github.com/andreyvit/diff"
)

var (
	create = flag.Bool("create", false, "create .golden files")
	update = flag.Bool("update", false, "update .golden files")
	yes    = true
	no     = false
)

type readerErrMock struct{}

func (readerErrMock) Read(p []byte) (int, error) {
	return 0, io.ErrNoProgress
}

type writerErrMock struct{}

func (w writerErrMock) Write(p []byte) (int, error) {
	return 0, io.ErrShortWrite
}

func _loadFixture(name string) ([]byte, error) {
	path := filepath.Join("testdata", name)
	return ioutil.ReadFile(path)
}

func loadInputFixture(t *testing.T, name string) []byte {
	fx, err := _loadFixture(name)
	if err != nil {
		t.Fatalf("error reading input fixture %s: %v", name, err)
	}
	return fx
}

func createGoldenFile(t *testing.T, name string) {
	path := filepath.Join("testdata", name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("error creating golden file %s: %v", name, err)
	}
	if err = f.Close(); err != nil {
		t.Fatalf("error closing golden file %s: %v", name, err)
	}
}

func updateGoldenFile(t *testing.T, name string, actual []byte) {
	path := filepath.Join("testdata", name)
	if err := ioutil.WriteFile(path, actual, 0644); err != nil {
		t.Fatalf("error updating golden file %s: %v", name, err)
	}
}

func loadGoldenFile(t *testing.T, name string, actual []byte) []byte {
	if *update {
		updateGoldenFile(t, name, actual)
	}
	g, err := _loadFixture(name)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok || !*create {
			t.Fatalf("error reading golden file %s: %v", name, err)
		}
		createGoldenFile(t, name)
		loadGoldenFile(t, name, actual)
	}
	return g
}

func TestPrettyPrinter_run_readError(t *testing.T) {
	pp := newPrettyPrinter(readerErrMock{}, &bytes.Buffer{}, nil)
	err := pp.run()

	if err == nil {
		t.Fatal("read should have failed")
	}
	if err != io.ErrNoProgress {
		t.Errorf("wanted %q got %v", io.ErrNoProgress, err)
	}
}

func TestPrettyPrinter_run_writeError(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdin.WriteString("foo\n")

	pp := newPrettyPrinter(stdin, new(writerErrMock), nil)
	err := pp.run()

	if err == nil {
		t.Fatal("write should have failed")
	}
	if err != io.ErrShortWrite {
		t.Errorf("wanted %q got %v", io.ErrShortWrite, err)
	}
}

func TestPrettyPrinter_processLine_writeError(t *testing.T) {
	pp := newPrettyPrinter(new(bytes.Buffer), new(writerErrMock), nil)
	input := "{\"version\":\"1.1\",\"host\":\"example.org\"," +
		"\"short_message\":\"foo\",\"timestamp\":1385053862.3072,\"level\":6}\n"
	err := pp.processLine([]byte(input))

	if err == nil {
		t.Fatal("write should have failed")
	}
	if err != io.ErrShortWrite {
		t.Errorf("wanted %q got %v", io.ErrShortWrite, err)
	}
}

func TestPrettyPrinter_run(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		out   string
		color bool
	}{
		{
			name: "valid",
			in:   "valid.input",
			out:  "valid.golden",
		},
		{
			name: "invalid",
			in:   "invalid.input",
			out:  "invalid.golden",
		},
		{
			name:  "colored",
			in:    "valid.input",
			out:   "colored.golden",
			color: true,
		},
	}

	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)

	defer func() { color.NoColor = true }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color.NoColor = !tt.color
			in := loadInputFixture(t, tt.in)
			stdin.Write(in)

			pp := newPrettyPrinter(stdin, stdout, time.UTC)
			if err := pp.run(); err != nil {
				t.Fatalf("unwanted error: %v", err)
			}

			actual := stdout.Bytes()
			expected := loadGoldenFile(t, tt.out, actual)

			if res := bytes.Compare(actual, expected); res != 0 {
				d := diff.LineDiff(string(expected), string(actual))
				t.Errorf("output not as expected:\n%v", d)
			}
		})
		stdin.Reset()
		stdout.Reset()
	}
}

func TestRun_version(t *testing.T) {
	if *versionFlag {
		t.Fatal("version flag should be false by default")
	}

	stdout := new(bytes.Buffer)

	version = "X.Y.Z"
	buildCommit = "640197df9b907efe9bfdf8ac2914b28a3ec9b8ef"
	buildTime = "2006-01-02T15:04:05Z"

	versionFlag = &yes
	defer func() { versionFlag = &no }()

	if err := run(new(bytes.Buffer), stdout); err != nil {
		t.Fatalf("unwanted error: %v", err)
	}

	actual := stdout.Bytes()
	expected := loadGoldenFile(t, "version.golden", actual)

	if res := bytes.Compare(actual, expected); res != 0 {
		d := diff.LineDiff(string(expected), string(actual))
		t.Errorf("output not as expected:\n%v", d)
	}
}

func TestRun_version_WriteError(t *testing.T) {
	versionFlag = &yes
	defer func() { versionFlag = &no }()

	err := run(new(bytes.Buffer), new(writerErrMock))

	if err == nil {
		t.Fatal("write should have failed")
	}
	if err != io.ErrShortWrite {
		t.Errorf("wanted %q got %v", io.ErrShortWrite, err)
	}
}

func TestRun_writeError(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdin.WriteString("foo\n")

	err := run(stdin, new(writerErrMock))

	if err == nil {
		t.Fatal("write should have failed")
	}
	if err != io.ErrShortWrite {
		t.Errorf("wanted %q got %v", io.ErrShortWrite, err)
	}
}

func TestRun_noColor(t *testing.T) {
	if *noColorFlag {
		t.Fatal("no-color flag is not false by default")
	}

	noColorFlag = &yes
	defer func() { noColorFlag = &no }()

	if err := run(new(bytes.Buffer), new(bytes.Buffer)); err != nil {
		t.Fatalf("unwanted error: %v", err)
	}
}
