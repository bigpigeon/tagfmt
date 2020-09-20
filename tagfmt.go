/*
 * Copyright 2020 bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 *
 */

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type tagFormatter struct {
	Err        error
	f          *ast.File
	fs         *token.FileSet
	needFormat [][]*ast.Field
}

func (s *tagFormatter) Scan() error {
	ast.Walk(s, s.f)
	return s.Err
}

func (s *tagFormatter) Execute() error {
	for _, fields := range s.needFormat {
		err := fieldsTagFormat(fields)
		if err != nil {
			s.Err = err
			return err
		}
	}
	return s.Err
}

func (s *tagFormatter) recordFields(fields *[]*ast.Field) {
	s.needFormat = append(s.needFormat, (*fields)[:])
	*fields = nil
}

func (s *tagFormatter) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.StructType:
		if n.Fields != nil {

			var multilineFields []*ast.Field
			var oneLineFields []*ast.Field

			for _, field := range n.Fields.List {

				fmt.Printf("name %s start %s \nend %s\n", field.Names[0].Name, s.fs.Position(field.Pos()), s.fs.Position(field.End()))
				line := s.fs.Position(field.Pos()).Line
				eline := s.fs.Position(field.End()).Line
				if eline-line > 0 {
					preELine := line
					if len(multilineFields) > 0 {
						preELine = s.fs.Position(multilineFields[len(multilineFields)-1].End()).Line
					}
					s.recordFields(&oneLineFields)
					if field.Tag == nil || line-preELine > 1 {
						s.recordFields(&multilineFields)

					}
					if field.Tag != nil {
						multilineFields = append(multilineFields, field)
					}
				} else {
					preLine := line
					if len(oneLineFields) > 0 {
						preLine = s.fs.Position(oneLineFields[len(oneLineFields)-1].Pos()).Line
					}

					s.recordFields(&multilineFields)

					if field.Tag == nil || line-preLine > 1 {
						s.recordFields(&oneLineFields)
					}
					if field.Tag != nil {
						oneLineFields = append(oneLineFields, field)
					}
				}

			}
			s.recordFields(&oneLineFields)
			s.recordFields(&multilineFields)

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

func newTagFmt(f *ast.File, fs *token.FileSet) *tagFormatter {
	s := &tagFormatter{fs: fs, f: f}
	return s
}
