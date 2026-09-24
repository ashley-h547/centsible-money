package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "amounts.txt")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return path
}

func TestRunPlainText(t *testing.T) {
	path := writeTempFile(t, "USD 19.99\nUSD 0.01\nJPY 1500\n")

	var stdout, stderr bytes.Buffer
	if err := run([]string{path}, &stdout, &stderr); err != nil {
		t.Fatalf("run() returned error: %v, stderr: %s", err, stderr.String())
	}

	want := "USD 20.00\nJPY 1500\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunSumOnly(t *testing.T) {
	path := writeTempFile(t, "USD 19.99\nUSD 0.01\n")

	var stdout, stderr bytes.Buffer
	if err := run([]string{"--sum-only", path}, &stdout, &stderr); err != nil {
		t.Fatalf("run() returned error: %v, stderr: %s", err, stderr.String())
	}

	want := "20.00\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunJSON(t *testing.T) {
	path := writeTempFile(t, "USD 19.99\nJPY 1500\n")

	var stdout, stderr bytes.Buffer
	if err := run([]string{"--json", path}, &stdout, &stderr); err != nil {
		t.Fatalf("run() returned error: %v, stderr: %s", err, stderr.String())
	}

	want := `[
  {
    "currency": "USD",
    "units": 1999,
    "formatted": "USD 19.99"
  },
  {
    "currency": "JPY",
    "units": 1500,
    "formatted": "JPY 1500"
  }
]
`
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunJSONSumOnly(t *testing.T) {
	path := writeTempFile(t, "USD 19.99\n")

	var stdout, stderr bytes.Buffer
	if err := run([]string{"--json", "--sum-only", path}, &stdout, &stderr); err != nil {
		t.Fatalf("run() returned error: %v, stderr: %s", err, stderr.String())
	}

	got := stdout.String()
	if strings.Contains(got, `"currency"`) {
		t.Errorf("stdout = %q, want no \"currency\" field under --sum-only", got)
	}
	if !strings.Contains(got, `"formatted": "19.99"`) {
		t.Errorf("stdout = %q, want formatted value with currency code stripped", got)
	}
}

func TestRunReportsParseErrors(t *testing.T) {
	path := writeTempFile(t, "XYZ 10.00\n")

	var stdout, stderr bytes.Buffer
	err := run([]string{path}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run() returned nil error, want an error for the unknown currency code")
	}
	if !strings.Contains(stderr.String(), "unknown currency code") {
		t.Errorf("stderr = %q, want it to mention the unknown currency code", stderr.String())
	}
	if stdout.String() != "" {
		t.Errorf("stdout = %q, want empty output when parsing failed", stdout.String())
	}
}
