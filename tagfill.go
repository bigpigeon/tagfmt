///*
// * Copyright 2020 bigpigeon. All rights reserved.
// * Use of this source code is governed by a MIT style
// * license that can be found in the LICENSE file.
// *
// */
//
package main

import (
	"go/ast"
	"go/token"
	"sort"
	"strings"
)

type tagFiller struct {
	Err error
	fs  *token.FileSet
}

func (s *tagFiller) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.StructType:
		if n.Fields != nil {
			keySet := map[string]struct{}{}
			var start int
			var end int
			var preFieldLine int
			for i, field := range n.Fields.List {
				line := s.fs.Position(field.Pos()).Line
				// If there are blank lines or nil field tag in the structure, reset
				if field.Tag == nil || preFieldLine+1 < line {
					fieldsTagFill(n.Fields.List[start:end], keySet)
					start = i
					end = i + 1
					keySet = map[string]struct{}{}
				}
				preFieldLine = line
				if field.Tag != nil {
					end = i + 1
					_, keyValues, err := ParseTag(field.Tag.Value)
					if err != nil {
						s.Err = err
						return nil
					}
					for _, kv := range keyValues {
						keySet[kv.Key] = struct{}{}
					}
				}
			}
			fieldsTagFill(n.Fields.List[start:], keySet)
		}
	}
	return s
}

func fieldsTagFill(fields []*ast.Field, keySet map[string]struct{}) {
	for _, f := range fields {
		missingKeySet := keySetClone(keySet)

		if f.Tag != nil {
			quote, keyValues, err := ParseTag(f.Tag.Value)
			if err != nil {
				// must be nil error
				panic(err)
			}
			var keyValueRaw []string
			for _, kv := range keyValues {
				keyValueRaw = append(keyValueRaw, kv.KeyValue)
				delete(missingKeySet, kv.Key)
			}
			missingKeys := make([]string, 0, len(missingKeySet))
			for k := range missingKeySet {
				missingKeys = append(missingKeys, k)
			}
			sort.Strings(missingKeys)

			for _, k := range missingKeys {
				keyValueRaw = append(keyValueRaw, k+":"+`""`)
			}

			f.Tag.Value = quote + strings.TrimRight(strings.Join(keyValueRaw, " "), " ") + quote
			f.Tag.ValuePos = 0
		}
	}
}

func keySetClone(keySet map[string]struct{}) map[string]struct{} {
	cl := make(map[string]struct{}, len(keySet))
	for k := range keySet {
		cl[k] = struct{}{}
	}
	return cl
}

func tagFill(f *ast.File, fs *token.FileSet) error {
	s := &tagFiller{fs: fs}
	ast.Walk(s, f)
	return s.Err
}
