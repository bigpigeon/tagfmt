// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Tagfmt formats  struct tag within Go programs.
It uses blanks for alignment.
tag must be in key:"value" pair format

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
	-f
		fill key and empty value for field

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

When invoke with -s tagfmt will sort struct tags by key.

	struct tag key example:
	struct User struct {
		Name     string `xml:"name" json:"name" yaml:"name"`
	}

	struct User struct {
		Name     string `json:"name" xml:"name" yaml:"name"`
	}

When invoke with -f tagfmt will fill missing key and empty value in group(group split by black line or field without tag)

	struct tag fill example:
	type User struct {
		Name     string `json:"name"`
		Password string `xml:"password"`
		EmptyTag string
		City     string `json:"group" xml:"group"`
		State    string `gorm:"type:varchar(64)" xml:"state"`
	}

	type User struct {
		Name     string `json:"name"    xml:""`
		Password string `xml:"password" json:""`
		EmptyTag string
		City     string `json:"group"            xml:"group" gorm:""`
		State    string `gorm:"type:varchar(64)" xml:"state" json:""`
	}
*/
package main
