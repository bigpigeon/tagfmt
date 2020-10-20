# tagfmt

Tagfmt formats struct tag within Go programs.

It uses blanks for alignment. tag must be in key:"value" pair format

tagfmt feature:

- align field tags in part 
- fill specified tag key to field
- sort field tags

## usage 
```
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
  -w    write result to (source) file instead of stdout

Debugging support:
	-cpuprofile filename
		Write cpu profile to the specified file.
```

## tag align 

align tag according to the longest tag's object of each column

```go 
// tagfmt .
package main
type Example struct {
	Data      string `xml:"data" yaml:"data"  json:"data"`
	OtherData string `xml:"other_data" json:"other_data:omitempty" yaml:"other_data"`
}
```
the result
```go
// tagfmt 
package main

type Example struct {
	Data      string `xml:"data"       yaml:"data"                 json:"data"`
	OtherData string `xml:"other_data" json:"other_data:omitempty" yaml:"other_data"`
}
```

space line or no tag's field will split this rule, just like struct field align by gofmt

```go 
//tagfmt
package main
type Example struct {
	Data string `xml:"data" yaml:"data,omitempty"  json:"data"`
	OtherData string `xml:"other_data" json:"other_data" yaml:"other_data"`

    NewLineData string `xml:"new_line_data" yaml:"new_line_data" json:"new_line_data"`
    NewLineOtherData string `xml:"new_line_other_data" yaml:"new_line_other_data"  json:"new_line_other_data"`
}
```

result

```go 
//tagfmt
package main

type Example struct {
	Data      string `xml:"data"       yaml:"data,omitempty" json:"data"`
	OtherData string `xml:"other_data" json:"other_data"     yaml:"other_data"`

	NewLineData      string `xml:"new_line_data"       yaml:"new_line_data"       json:"new_line_data"`
	NewLineOtherData string `xml:"new_line_other_data" yaml:"new_line_other_data" json:"new_line_other_data"`
}
```

## tag fill

tag fill can fill specified key to field tag



## tag sort 

```
//tagfmt -s
package main
type Example struct {
	Data string `xml:"data" yaml:"data"  json:"data"  `
}
```

after sort 

```
//tagfmt -s 
package main

type Example struct {
	Data string `json:"data" xml:"data" yaml:"data"`
}
```
