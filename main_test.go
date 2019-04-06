package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestLevelToName(t *testing.T) {
	wanted := map[int]string{
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
		if v != wanted[k] {
			t.Errorf("invalid level name, wanted %v got %s", wanted[k], v)
		}
	}
}

func TestVersionInfo(t *testing.T) {
	Version = "0.0.0"
	BuildCommit = "640197df9b907efe9bfdf8ac2914b28a3ec9b8ef"
	BuildTime = "2019-03-30T12:48:27Z"

	got := versionInfo().String()
	wanted := "\n" +
		"          Version:  0.0.0\n" +
		"Build Commit Hash:  640197df9b907efe9bfdf8ac2914b28a3ec9b8ef\n" +
		"       Build Time:  2019-03-30T12:48:27Z\n" +
		"\n"
	if got != wanted {
		t.Errorf("wanted %q got '%v'", wanted, got)
	}
}

func TestPrettyPrinter_run(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		out  []string
	}{
		{"simple", []string{"foo", "bar"}, []string{"foo", "bar"}},
		{"skipBlanks", []string{"foo", "", "bar"}, []string{"foo", "bar"}},
	}

	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, i := range tt.in {
				stdin.WriteString(i)
				stdin.WriteString("\n")
			}

			pp := newPrettyPrinter(stdin, stdout)
			if err := pp.run(); err != nil {
				t.Fatalf("unwanted error: %v", err)
			}

			out := strings.Split(stdout.String(), "\n")
			out = out[:len(out)-1]
			for i, o := range out {
				if o != tt.out[i] {
					t.Errorf("wanted %s got %s", out[i], o)
				}
			}

			stdin.Reset()
			stdout.Reset()
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
		t.Errorf("wanted %q got %v", io.ErrNoProgress, err)
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
		t.Errorf("wanted %q got %v", io.ErrShortWrite, err)
	}
}
