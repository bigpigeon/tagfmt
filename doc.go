// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Tagfmt formats  struct tag within Go programs.
It uses blanks for alignment.
Alignment assumes that an editor is using a fixed-width font.

Usage:
	tagfmt [flags] [path ...]

The flags are:
	-cpuprofile string
		write cpu profile to this file
	-d
		display diffs instead of rewriting files
	-e
		report all errors (not just the first 10 on different lines)
	-l
		list files whose formatting differs from tagfmt's
	-s
		sort struct tag by key
	-w
		write result to (source) file instead of stdout

Debugging support:
	-cpuprofile filename
		Write cpu profile to the specified file.


Examples

	struct tag format example:
	struct User struct {
		Name     string `json:"name" xml:"name" yaml:"name"`
		Password string `json:"password" xml:"password" yaml:"password"`
	}

	struct User struct {
		Name     string `json:"name"     xml:"name"     yaml:"name"    `
		Password string `json:"password" xml:"password" yaml:"password"`
	}

When invoke with -s gofmt will sort struct tags by key.

	struct tag key example:
	struct User struct {
		Name     string `xml:"name" json:"name" yaml:"name"`
	}

	struct User struct {
		Name     string `json:"name" xml:"name" yaml:"name"`
	}
*/
package main
