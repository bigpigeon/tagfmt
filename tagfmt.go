/*
 * Copyright 2020 bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 *
 */

package main

import (
	"go/ast"
	"go/token"
	"strings"
)

type tagFormatter struct {
	Err error
	fs  *token.FileSet
}

func (s *tagFormatter) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.StructType:
		if n.Fields != nil {
			var start int
			var end int
			var preFieldLine int
			for i, field := range n.Fields.List {
				line := s.fs.Position(field.Pos()).Line
				// If there are blank lines or nil field tag in the structure, reset
				if field.Tag == nil || preFieldLine+1 < line {
					err := fieldsTagFormat(n.Fields.List[start:end])
					if err != nil {
						s.Err = err
						return nil
					}
					start = i
					if field.Tag == nil {
						start++
					}
					end = i + 1
				}
				preFieldLine = line
				end = i + 1

			}
			err := fieldsTagFormat(n.Fields.List[start:])
			if err != nil {
				s.Err = err
				return nil
			}
		}
	}
	return s
}

func fieldsTagFormat(fields []*ast.Field) error {
	var longestList []int
	for _, f := range fields {
		_, keyValues, err := ParseTag(f.Tag.Value)
		if err != nil {
			return err
		}
		for i, kv := range keyValues {
			if i >= len(longestList) {
				longestList = append(longestList, 0)
			}
			longestList[i] = max(len(kv.KeyValue), longestList[i])
		}
	}

	for _, f := range fields {
		if f.Tag != nil {
			quote, keyValues, err := ParseTag(f.Tag.Value)
			if err != nil {
				// must be nil error
				panic(err)
			}
			var keyValueRaw []string
			for i, kv := range keyValues {
				keyValueRaw = append(keyValueRaw, kv.KeyValue+strings.Repeat(" ", longestList[i]-len(kv.KeyValue)))
			}

			f.Tag.Value = quote + strings.TrimRight(strings.Join(keyValueRaw, " "), " ") + quote
			f.Tag.ValuePos = 0
		}
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func tagFmt(f *ast.File, fs *token.FileSet) error {
	s := &tagFormatter{fs: fs}
	ast.Walk(s, f)
	return s.Err
}
