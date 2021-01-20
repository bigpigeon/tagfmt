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
	"sort"
	"strings"
)

type tagSorterWeightKey struct {
	Weight int
	Key    string
}

type tagSorter struct {
	f       *ast.File
	fs      *token.FileSet
	Err     error
	order   []string
	weights map[string]int
	fields  []*ast.Field
}

func (s *tagSorter) Scan() error {
	ast.Walk(s, s.f)
	return s.Err
}

func (s *tagSorter) Execute() error {
	for _, field := range s.fields {
		err := sortField(field, s.order, s.weights)
		if err != nil {
			s.Err = err
			return err
		}
	}
	return s.Err
}

func (s *tagSorter) Visit(node ast.Node) ast.Visitor {
	cmap := ast.NewCommentMap(s.fs, node, s.f.Comments)
	visit := newTopVisit(cmap, s.executor)
	return visit.Visit(node)
}

func (s *tagSorter) executor(name string, comments []*ast.CommentGroup, n *ast.StructType) {
	if n.Fields != nil {
		for _, field := range n.Fields.List {
			if fieldFilter(getFieldName(field)) && field.Tag != nil {
				s.fields = append(s.fields, field)
			}
		}
	}
}

func sortField(field *ast.Field, order []string, weight map[string]int) error {
	quote, keyValues, err := ParseTag(field.Tag.Value)
	if err != nil {
		return err
	}
	sort.Slice(keyValues, func(i, j int) bool {
		iKey := keyValues[i].Key
		jKey := keyValues[j].Key
		if weight[iKey] > weight[jKey] {
			return true
		} else if weight[iKey] < weight[jKey] {
			return false
		}
		for _, o := range order {
			if iKey == o {
				return true
			} else if jKey == o {
				return false
			}
		}

		return iKey < jKey
	})
	var keyValuesRaw []string
	for _, kv := range keyValues {
		keyValuesRaw = append(keyValuesRaw, kv.String())
	}

	field.Tag.Value = quote + strings.Join(keyValuesRaw, " ") + quote
	field.Tag.ValuePos = 0
	return nil
}

func newTagSort(f *ast.File, fs *token.FileSet, order []string, weights map[string]int) *tagSorter {
	s := &tagSorter{f: f, order: order, fs: fs, weights: weights}

	return s
}
