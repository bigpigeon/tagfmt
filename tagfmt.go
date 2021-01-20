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
	"unicode/utf8"
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

func (s *tagFormatter) recordFields(fwt []*ast.Field) {
	if len(fwt) != 0 {
		s.needFormat = append(s.needFormat, fwt)
	}
}

func getFieldName(node *ast.Field) string {
	if len(node.Names) > 0 {
		return node.Names[0].Name
	}

	return ""
}

func getFieldOrTypeName(node *ast.Field) string {
	if len(node.Names) > 0 {
		return node.Names[0].Name
	}
	//if ident, ok := node.Type.(*ast.Ident); ok {
	//	return ident.Name
	//}
	return ""
}

type tagFormatterFields struct {
	multiline []*ast.Field
	oneline   []*ast.Field
	anonymous []*ast.Field
	s         *tagFormatter
}

func (fields *tagFormatterFields) reset(f *tagFormatter) {
	f.recordFields(fields.multiline)
	fields.multiline = nil
	f.recordFields(fields.oneline)
	fields.oneline = nil
	f.recordFields(fields.anonymous)
	fields.anonymous = nil
}

func (s *tagFormatter) executor(name string, comments []*ast.CommentGroup, n *ast.StructType) {
	if n.Fields != nil {
		var ffields tagFormatterFields

		if len(n.Fields.List) == 0 {
			return
		}
		preMultiELine := -1
		preEline := -1
		preAnonymousELine := -1
		for _, field := range n.Fields.List {
			fieldName := getFieldOrTypeName(field)
			if field.Tag == nil || fieldFilter(fieldName) == false {
				ffields.reset(s)
				continue
			}

			line := s.fs.Position(field.Pos()).Line
			eline := s.fs.Position(field.End()).Line
			// the one way to distinguish the field with multiline anonymous struct and others
			if len(field.Names) == 0 {
				if line-preAnonymousELine > 1 {
					ffields.reset(s)
				}
				ffields.anonymous = append(ffields.anonymous, field)
				preAnonymousELine = eline
			} else if eline-line > 0 {
				if line-preMultiELine > 1 {
					ffields.reset(s)
				}
				ffields.multiline = append(ffields.multiline, field)
				preMultiELine = eline
			} else {
				if line-preEline > 1 {
					ffields.reset(s)
				}
				ffields.oneline = append(ffields.oneline, field)
				preEline = eline
			}
		}
		ffields.reset(s)
	}
}

func (s *tagFormatter) Visit(node ast.Node) ast.Visitor {
	cmap := ast.NewCommentMap(s.fs, node, s.f.Comments)
	visit := newTopVisit(cmap, s.executor)
	return visit.Visit(node)
}

func fieldsTagFormat(fields []*ast.Field) error {
	var longestList []int
	for _, field := range fields {
		_, keyWords, err := ParseTag(field.Tag.Value)
		if err != nil {
			return err
		}
		for i, kv := range keyWords {
			if i >= len(longestList) {
				longestList = append(longestList, 0)
			}
			kvLen := utf8.RuneCountInString(kv.String())
			longestList[i] = max(kvLen, longestList[i])
		}
	}

	for _, field := range fields {
		quote, keyWords, err := ParseTag(field.Tag.Value)
		if err != nil {
			return err
		}
		var keyValueRaw []string
		for i, kv := range keyWords {
			kvLen := utf8.RuneCountInString(kv.String())
			keyValueRaw = append(keyValueRaw, kv.String()+strings.Repeat(" ", longestList[i]-kvLen))
		}

		field.Tag.Value = quote + strings.TrimRight(strings.Join(keyValueRaw, " "), " ") + quote
		field.Tag.ValuePos = 0
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
