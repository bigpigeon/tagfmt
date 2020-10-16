# tagfmt

Tagfmt formats struct tag within Go programs.

It uses blanks for alignment. tag must be in key:"value" pair format

tagfmt feature:

- align field tags in part 
- fill specified tag key to field
- sort field tags

## tag align 

```
// tagfmt .
package main
type Example struct {
	Data      string `xml:"data" yaml:"data"  json:"data"`
	OtherData string `xml:"other_data" json:"other_data:omitempty" yaml:"other_data"`
}
```
the result
```

```