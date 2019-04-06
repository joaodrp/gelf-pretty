package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestVersionInfo(t *testing.T) {
	// FIXME: write proper tests
	printVersion()
}

func TestLevelToNameMap(t *testing.T) {
	want := map[int]string{
		0: "EMERGENCY",
		1: "ALERT",
		2: "CRITICAL",
		3: "ERROR",
		4: "WARNING",
		5: "NOTICE",
		6: "INFO",
		7: "DEBUG",
	}
	for k, v := range levelToName {
		if v != want[k] {
			t.Errorf("invalid level name, expected %v got %s.", want[k], v)
		}
	}
}

var runTests = []struct {
	name string
	in   []string
	out  []string
}{
	{"simple", []string{"foo", "bar"}, []string{"foo", "bar"}},
	{"skipBlanks", []string{"foo", "", "bar"}, []string{"foo", "bar"}},
}

func TestPrettyPrinter_run(t *testing.T) {
	for _, tt := range runTests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := new(bytes.Buffer)
			stdout := new(bytes.Buffer)

			for _, i := range tt.in {
				stdin.WriteString(i)
				stdin.WriteString("\n")
			}

			pp := newPrettyPrinter(stdin, stdout)

			if err := pp.run(); err != nil {
				t.Fatalf("unexpected error: %v.", err)
			}

			out := strings.Split(stdout.String(), "\n")
			out = out[:len(out)-1]
			for i, o := range out {
				if o != tt.out[i] {
					t.Errorf("expected %s got %s.", out[i], o)
				}
			}

		})
	}
}

type readerErrMock struct{}

func (readerErrMock) Read(p []byte) (int, error) {
	return 0, io.ErrNoProgress
}

func TestPrettyPrinter_run_readError(t *testing.T) {
	pp := newPrettyPrinter(readerErrMock{}, &bytes.Buffer{})
	err := pp.run()

	if err == nil {
		t.Fatal("read should have failed")
	}
	if err != io.ErrNoProgress {
		t.Errorf("expected %q got %v", io.ErrNoProgress, err)
	}
}

type writerErrMock struct{}

func (w writerErrMock) Write(p []byte) (int, error) {
	return 0, io.ErrShortWrite
}

func TestPrettyPrinter_run_writeError(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdin.WriteString("foo\n")

	pp := newPrettyPrinter(stdin, &writerErrMock{})
	err := pp.run()

	if err == nil {
		t.Fatal("write should have failed")
	}
	if err != io.ErrShortWrite {
		t.Errorf("expected %q got %v", io.ErrShortWrite, err)
	}
}
