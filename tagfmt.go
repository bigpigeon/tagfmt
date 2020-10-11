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
			if len(n.Fields.List) == 0 {
				return s
			}
			preMultiELine := -1
			preEline := -1
			for _, field := range n.Fields.List {
				if fieldFilter(field.Names[0].Name) == false {
					continue
				}
				line := s.fs.Position(field.Pos()).Line
				eline := s.fs.Position(field.End()).Line
				// the one way to distinguish the field with multiline anonymous struct and others
				if eline-line > 0 {
					s.recordFields(&oneLineFields)
					if field.Tag == nil || line-preMultiELine > 1 {
						s.recordFields(&multilineFields)
					}
					if field.Tag != nil {
						multilineFields = append(multilineFields, field)
					}
					preMultiELine = eline
				} else {
					s.recordFields(&multilineFields)
					if field.Tag == nil || line-preEline > 1 {
						s.recordFields(&oneLineFields)
					}
					if field.Tag != nil {
						oneLineFields = append(oneLineFields, field)
					}
					preEline = eline
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
			longestList[i] = max(len(kv.String()), longestList[i])
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
				keyValueRaw = append(keyValueRaw, kv.String()+strings.Repeat(" ", longestList[i]-len(kv.String())))
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
