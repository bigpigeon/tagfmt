// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was copied from the src/cmd/gofmt/gofmt_test.go

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"text/scanner"
)

var update = flag.Bool("update", false, "update .golden files")

// gofmtFlags looks for a comment of the form
//
//	//gofmt flags
//
// within the first maxLines lines of the given file,
// and returns the flags string, if any. Otherwise it
// returns the empty string.
func gofmtFlags(filename string, maxLines int) string {
	f, err := os.Open(filename)
	if err != nil {
		return "" // ignore errors - they will be found later
	}
	defer f.Close()

	// initialize scanner
	var s scanner.Scanner
	s.Init(f)
	s.Error = func(*scanner.Scanner, string) {}       // ignore errors
	s.Mode = scanner.GoTokens &^ scanner.SkipComments // want comments

	// look for //gofmt comment
	for s.Line <= maxLines {
		switch s.Scan() {
		case scanner.Comment:
			const prefix = "//tagfmt "
			if t := s.TokenText(); strings.HasPrefix(t, prefix) {
				return strings.TrimSpace(t[len(prefix):])
			}
		case scanner.EOF:
			return ""
		}

	}

	return ""
}

// gofmtError looks for a comment of the form
//
//	//error: message
//
// within the first maxLines lines of the given file,
// and returns the error string, if any. Otherwise it
// returns the empty string.
func gofmtError(filename string, maxLines int) string {
	f, err := os.Open(filename)
	if err != nil {
		return "" // ignore errors - they will be found later
	}
	defer f.Close()

	// initialize scanner
	var s scanner.Scanner
	s.Init(f)
	s.Error = func(*scanner.Scanner, string) {}       // ignore errors
	s.Mode = scanner.GoTokens &^ scanner.SkipComments // want comments

	// look for //gofmt comment
	for s.Line <= maxLines {
		switch s.Scan() {
		case scanner.Comment:
			const prefix = "//error: "
			if t := s.TokenText(); strings.HasPrefix(t, prefix) {
				return strings.TrimSpace(t[len(prefix):])
			}
		case scanner.EOF:
			return ""
		}

	}
	return ""
}

func runTest(t *testing.T, in, out string) {
	// process flags
	stdin := false
	resetFlags()
	var nextVal func(s string)
	for _, flag := range strings.Split(gofmtFlags(in, 20), " ") {
		if nextVal != nil {
			nextVal(flag)
			nextVal = nil
			continue
		}

		switch flag {
		case "":
			// no flags
		case "-stdin":
			// fake flag - pretend input is from stdin
			stdin = true
		case "-s":
			*tagSort = true
		case "-f":
			nextVal = func(s string) {
				var err error
				*fill, err = strconv.Unquote(s)
				if err != nil {
					panic("err: " + err.Error() + " str: " + s)
				}
			}
		case "-p":
			nextVal = func(s string) {
				var err error
				*pattern, err = strconv.Unquote(s)
				if err != nil {
					panic(err)
				}
			}
		case "-P":
			nextVal = func(s string) {
				var err error
				*inversePattern, err = strconv.Unquote(s)
				if err != nil {
					panic(err)
				}
			}
		case "-sp":
			nextVal = func(s string) {
				var err error
				*structPattern, err = strconv.Unquote(s)
				if err != nil {
					panic(err)
				}
			}
		case "-sP":
			nextVal = func(s string) {
				var err error
				*inverseStructPattern, err = strconv.Unquote(s)
				if err != nil {
					panic(err)
				}
			}
		case "-so":
			nextVal = func(s string) {
				var err error
				*tagSortOrder, err = strconv.Unquote(s)
				if err != nil {
					panic(err)
				}
			}
		case "-sw":
			nextVal = func(s string) {
				var err error
				*tagSortWeight, err = strconv.Unquote(s)
				if err != nil {
					panic(err)
				}
			}
		default:
			t.Errorf("unrecognized flag name: %s", flag)
		}
	}

	initParserMode()
	mustError := strings.TrimSpace(gofmtError(in, 20))

	var buf bytes.Buffer
	err := processFile(in, nil, &buf, stdin)
	if err != nil {
		if mustError != "" {
			errStr := strings.TrimSpace(strings.Replace(err.Error(), "\n", "", -1))
			if mustError != errStr {
				t.Error("expected got err \"" + mustError + "\", in fact got err \"" + errStr + "\"")

			}
		} else {
			t.Error(err)
		}
		return
	} else if mustError != "" {
		t.Error("file " + in + " must return error")
	}

	expected, err := ioutil.ReadFile(out)
	if err != nil {
		t.Error(err)
		return
	}

	if got := buf.Bytes(); !bytes.Equal(got, expected) {
		if *update {
			if in != out {
				if err := ioutil.WriteFile(out, got, 0666); err != nil {
					t.Error(err)
				}
				return
			}
			// in == out: don't accidentally destroy input
			t.Errorf("WARNING: -update did not rewrite input file %s", in)
		}

		t.Errorf("(gofmt %s) != %s (see %s.gofmt)", in, out, in)
		d, err := diff(expected, got, in)
		if err == nil {
			t.Errorf("%s", d)
		}
		if err := ioutil.WriteFile(in+".gofmt", got, 0666); err != nil {
			t.Error(err)
		}
	}
}

// TestRewrite processes testdata/*.input files and compares them to the
// corresponding testdata/*.golden files. The gofmt flags used to process
// a file must be provided via a comment of the form
//
//	//gofmt flags
//
// in the processed file within the first 20 lines, if any.
func TestRewrite(t *testing.T) {
	// determine input files
	match, err := filepath.Glob("testdata/*.input")
	if err != nil {
		t.Fatal(err)
	}

	// add larger examples
	match = append(match, "gofmt.go", "gofmt_test.go")

	for _, in := range match {
		out := in // for files where input and output are identical
		if strings.HasSuffix(in, ".input") {
			out = in[:len(in)-len(".input")] + ".golden"
		}
		runTest(t, in, out)
		if in != out {
			// Check idempotence.
			runTest(t, out, out)
		}
	}
}

func TestDiff(t *testing.T) {
	if _, err := exec.LookPath("diff"); err != nil {
		t.Skipf("skip test on %s: diff command is required", runtime.GOOS)
	}
	in := []byte("first\nsecond\n")
	out := []byte("first\nthird\n")
	filename := "difftest.txt"
	b, err := diff(in, out, filename)
	if err != nil {
		t.Fatal(err)
	}

	if runtime.GOOS == "windows" {
		b = bytes.ReplaceAll(b, []byte{'\r', '\n'}, []byte{'\n'})
	}

	bs := bytes.SplitN(b, []byte{'\n'}, 3)
	line0, line1 := bs[0], bs[1]

	if prefix := "--- difftest.txt.orig"; !bytes.HasPrefix(line0, []byte(prefix)) {
		t.Errorf("diff: first line should start with `%s`\ngot: %s", prefix, line0)
	}

	if prefix := "+++ difftest.txt"; !bytes.HasPrefix(line1, []byte(prefix)) {
		t.Errorf("diff: second line should start with `%s`\ngot: %s", prefix, line1)
	}

	want := `@@ -1,2 +1,2 @@
 first
-second
+third
`

	if got := string(bs[2]); got != want {
		t.Errorf("diff: got:\n%s\nwant:\n%s", got, want)
	}
}

func TestReplaceTempFilename(t *testing.T) {
	diff := []byte(`--- /tmp/tmpfile1	2017-02-08 00:53:26.175105619 +0900
+++ /tmp/tmpfile2	2017-02-08 00:53:38.415151275 +0900
@@ -1,2 +1,2 @@
 first
-second
+third
`)
	want := []byte(`--- path/to/file.go.orig	2017-02-08 00:53:26.175105619 +0900
+++ path/to/file.go	2017-02-08 00:53:38.415151275 +0900
@@ -1,2 +1,2 @@
 first
-second
+third
`)
	// Check path in diff output is always slash regardless of the
	// os.PathSeparator (`/` or `\`).
	sep := string(os.PathSeparator)
	filename := strings.Join([]string{"path", "to", "file.go"}, sep)
	got, err := replaceTempFilename(diff, filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("os.PathSeparator='%s': replacedDiff:\ngot:\n%s\nwant:\n%s", sep, got, want)
	}
}
