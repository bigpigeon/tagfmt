// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Tagfmt formats struct tag within Go programs.
It uses blanks for alignment.
tag must be in key:"value" pair format

usage: tagfmt [flags] [path ...]
  -P string
        field name with inverse regular expression pattern
  -a    align with nearby field's tag (default true)
  -cpuprofile string
        write cpu profile to this file
  -d    display diffs instead of rewriting files
  -e    report all errors (not just the first 10 on different lines)
  -f string
        fill key and value for field e.g json=lower(_val)|yaml=snake(_val)
  -l    list files whose formatting differs from tagfmt's
  -p string
        field name with regular expression pattern (default ".*")
  -s    sort struct tag by key
  -sP string
        struct name with inverse regular expression pattern
  -so string
        sort struct tag keys order e.g json|yaml|desc
  -sp string
        struct name with regular expression pattern (default ".*")
  -w    write result to (source) file instead of stdout



Debugging support:
	-cpuprofile filename
		Write cpu profile to the specified file.


Examples

	struct tag format example:
	//tagfmt
	struct User struct {
		Name     string `json:"name" xml:"name" yaml:"name"`
		Password string `json:"password" xml:"password" yaml:"password"`
	}
	// after format
	struct User struct {
		Name     string `json:"name"     xml:"name"     yaml:"name"    `
		Password string `json:"password" xml:"password" yaml:"password"`
	}

When invoke with -s tagfmt will sort struct tags by key.

	struct tag key example:
	//tagfmt -s
	struct User struct {
		Name     string `xml:"name" json:"name" yaml:"name"`
	}
	// after format
	struct User struct {
		Name     string `json:"name" xml:"name" yaml:"name"`
	}

When invoke with -so <order> and -s will sort struct tags by your custom <order>

	//tagfmt -s -so "json|yaml|desc"
	package main
	type Example struct {
		Data string `desc:"some inuse data" yaml:"data" json:"data" `
	}

	type Example struct {
		Data string `json:"data" yaml:"data" desc:"some inuse data"`
	}

When invoke with -sw <weight> and -s will sort struct tags by your custom <weight>

	//tagfmt -s -sw "json=2|yaml=1|toml=1|desc=-1"
	package main
	type Example struct {
		Data string `desc:"some inuse data" yaml:"data" toml:"data" binding:"required" json:"data" `
	}

	package main

	type Example struct {
		Data string `json:"data" toml:"data" yaml:"data" binding:"required" desc:"some inuse data"`
	}


When invoke with -f "*" tagfmt will fill missing key and empty value in group(group split by black line or field without tag)

	struct tag fill example:
	//tagfmt -f "*"
	type User struct {
		Name     string `json:"name"`
		Password string `xml:"password"`
		EmptyTag string
		City     string `json:"group" xml:"group"`
		State    string `gorm:"type:varchar(64)" xml:"state"`
	}
	// after format
	type User struct {
		Name     string `json:"name"    xml:""`
		Password string `xml:"password" json:""`
		EmptyTag string
		City     string `json:"group"            xml:"group" gorm:""`
		State    string `gorm:"type:varchar(64)" xml:"state" json:""`
	}

You also can only fill "json" tag key and field name as its value

    struct tag fill example:
	//tagfmt -f "json=:field"

	type Order struct {
		ID  string ``
		Tag string ``
		Fee float32 ``
	}
	// after format
	type Order struct {
		ID  string  `json:"ID"`
		Tag string  `json:"Tag"`
		Fee float32 `json:"Fee"`
	}


use `// tagfill: [key1 key2]` to filter below struct requires key

	struct tag fill example:
	//tagfmt -f "json=snake(:tag)|yaml=lower_camel(:tag)|bson=lower_camel(:tag)|toml=upper_camel(:tag)"

	package main
	// tagfill: toml yaml
	type OrderConfig struct {
		Name     string ``
		UserName string ``
		Pay      int    ``
	}
	// tagfill: json bson
	type OrderDetail struct {
		ID       string ``
		UserName string ``
		Pay      int    ``
	}


	//after format

	package main

	// tagfill: toml yaml
	type OrderConfig struct {
		Name     string `toml:"" yaml:""`
		UserName string `toml:"" yaml:""`
		Pay      int    `toml:"" yaml:""`
	}

	// tagfill: json bson
	type OrderDetail struct {
		ID       string `bson:"" json:""`
		UserName string `bson:"" json:""`
		Pay      int    `bson:"" json:""`
	}


	fill rule are rich and flexible here is example about fill json key and snake converted field name as its value, final keep it's origin extra tag
	struct tag fill example:
	//tagfmt -f "json=snake(:field)+:tag_extra"

	package main
	type OrderDetail struct {
		ID       string   `json:",omitempty"`
		UserName string   `json:",omitempty"`
		OrderID  string   `json:",omitempty"`
		Callback string   ``
		Address  []string ``
	}
	// after format
	type OrderDetail struct {
		ID       string   `json:"id,omitempty"`
		UserName string   `json:"user_name,omitempty"`
		OrderID  string   `json:"order_id,omitempty"`
		Callback string   `json:"callback"`
		Address  []string `json:"address"`
	}

	fill rule:
		multiple key rule split with '|'
		<key>[=<function or placehold_val or string>[+ <function or placehold_val or string> ]]
		'*' is special key, it will fill missing key and empty value in group(group split by black line or field without tag)

	fill rule functions:
		upper(s string) // a-z to A-Z
		lower(s string) // A-Z to a-z
		snake(s string) // convert upper_camel/lower_camel word to snake case
		upper_camel(s string) // convert snake case/lower camel case to upper camel case
		lower_camel(s string) // convert upper camel case/snake case to lower camel case
		or(s string, s string) // return return first params if it's not zero,else return the second

	fill rule placehold value:
		:field // replace with struct field name
		:tag   // replace with  struct field existed tag's value
		:tag_basic // replace with field existed tag's basic value (the value before the first ',' )
		:tag_extra // replace with field existed tag's extra data (the value after the first ',' )

	fill Concatenated string
		fill rule also support use '+' to concatenated string
		//tagfmt -f "json=snake(:tag_basic)+',omitempty'"

		type OrderDetail struct {
			ID       string   `json:"id"`
			UserName string   `json:"user_name"`
			OrderID  string   `json:"order_id"`
			Callback string   `json:"callback"`
			Address  []string `json:"address"`
		}

		type OrderDetail struct {
			ID       string   `json:"id,omitempty"`
			UserName string   `json:"user_name,omitempty"`
			OrderID  string   `json:"order_id,omitempty"`
			Callback string   `json:"callback,omitempty"`
			Address  []string `json:"address,omitempty"`
		}

*/
package main
