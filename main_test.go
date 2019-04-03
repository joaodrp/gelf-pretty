package main

import "testing"

func TestPrintVersion(t *testing.T) {
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
			t.Errorf("Invalid level name, want: %v, got: %s.", want[k], v)
		}
	}
}
