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
	Err error
	fs  *token.FileSet
}

func (s *tagFormatter) resetFields(fields *[]*ast.Field) (hasError bool) {
	err := fieldsTagFormat(*fields)
	if err != nil {
		s.Err = err
		return true
	}
	*fields = nil
	return false
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
					if s.resetFields(&oneLineFields) {
						return nil
					}
					if field.Tag == nil || line-preELine > 1 {
						if s.resetFields(&multilineFields) {
							return nil
						}
					}
					if field.Tag != nil {
						multilineFields = append(multilineFields, field)
					}
				} else {
					preLine := line
					if len(oneLineFields) > 0 {
						preLine = s.fs.Position(oneLineFields[len(oneLineFields)-1].Pos()).Line
					}

					if s.resetFields(&multilineFields) {
						return nil
					}

					if field.Tag == nil || line-preLine > 1 {
						if s.resetFields(&oneLineFields) {
							return nil
						}
					}
					if field.Tag != nil {
						oneLineFields = append(oneLineFields, field)
					}
				}

			}
			if s.resetFields(&oneLineFields) {
				return nil
			}
			if s.resetFields(&multilineFields) {
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
