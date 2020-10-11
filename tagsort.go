/*
 * Copyright 2020 bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 *
 */

package main

import (
	"errors"
	"go/ast"
	"sort"
	"strings"
)

var ErrInvalidTag = errors.New("Invalid tag ")

type tagSorter struct {
	f      *ast.File
	Err    error
	fields []*ast.Field
}

func (s *tagSorter) Scan() error {
	ast.Walk(s, s.f)
	return s.Err
}

func (s *tagSorter) Execute() error {
	for _, field := range s.fields {
		err := sortField(field)
		if err != nil {
			s.Err = err
			return err
		}
	}
	return s.Err
}

func (s *tagSorter) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.StructType:
		if n.Fields != nil {
			for _, field := range n.Fields.List {
				if field.Tag != nil {
					s.fields = append(s.fields, field)
				}
			}
		}
	}
	return s
}

func sortField(field *ast.Field) error {
	quote, keyValues, err := ParseTag(field.Tag.Value)
	if err != nil {
		return err
	}
	sort.Slice(keyValues, func(i, j int) bool {
		return keyValues[i].Key < keyValues[j].Key
	})
	var keyValuesRaw []string
	for _, kv := range keyValues {
		if fieldFilter(field.Names[0].Name) == false {
			continue
		}
		keyValuesRaw = append(keyValuesRaw, kv.String())
	}

	field.Tag.Value = quote + strings.Join(keyValuesRaw, " ") + quote
	field.Tag.ValuePos = 0
	return nil
}

func newTagSort(f *ast.File) *tagSorter {
	s := &tagSorter{f: f}

	return s
}
