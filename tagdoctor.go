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
)

const tagDockerMaxErr = 5

type tagDockerErr []error

func (e tagDockerErr) Error() string {
	s := "detect error: \n"
	for _, _e := range e {
		s += "    " + _e.Error() + "\n"
	}
	return s
}

type tagDoctor struct {
	f   *ast.File
	fs  *token.FileSet
	Err tagDockerErr
}

func (s *tagDoctor) Visit(node ast.Node) ast.Visitor {
	visit := toyVisit{executor: s.executor}
	return visit.Visit(node)
}

func (t *tagDoctor) executor(name string, n *ast.StructType) {
	if n.Fields != nil {
		for _, field := range n.Fields.List {
			fieldName := getFieldOrTypeName(field)
			if fieldFilter(fieldName) == false {
				continue
			}
			if field.Tag != nil {
				_, _, err := ParseTag(field.Tag.Value)
				if err != nil {
					if len(t.Err) < tagDockerMaxErr {
						t.Err = append(t.Err, NewAstError(t.fs, field.Tag, err))
					}
				}
			}
		}
	}
	return
}

func (t *tagDoctor) Scan() error {
	ast.Walk(t, t.f)
	if len(t.Err) != 0 {
		return t.Err
	}
	return nil
}

func (t *tagDoctor) Execute() error {
	return nil
}
