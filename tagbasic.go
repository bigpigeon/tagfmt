/*
 * Copyright 2020 bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 *
 */

package main

import "go/ast"

type toyVisitExecutor func(name string, comments []*ast.CommentGroup, n *ast.StructType)

type toyVisit struct {
	executor toyVisitExecutor
	cmap     ast.CommentMap
	Comments []*ast.CommentGroup
}

func newTopVisit(cmap ast.CommentMap, executor toyVisitExecutor) *toyVisit {
	return &toyVisit{
		cmap:     cmap,
		executor: executor,
	}
}

func (s *toyVisit) Copy() *toyVisit {
	comments := make([]*ast.CommentGroup, len(s.Comments))
	copy(comments, s.Comments)
	return &toyVisit{
		executor: s.executor,
		cmap:     s.cmap,
		Comments: comments,
	}
}

func (s *toyVisit) WithComments(comments []*ast.CommentGroup) *toyVisit {
	return &toyVisit{
		executor: s.executor,
		cmap:     s.cmap,
		Comments: comments,
	}
}

func (s *toyVisit) rangeField(fields *ast.FieldList) {
	if fields != nil {
		for _, f := range fields.List {
			if _struct, ok := f.Type.(*ast.StructType); ok {
				s.executor("", s.Comments, _struct)
				s.rangeField(_struct.Fields)
			}
		}
	}
}

func (s *toyVisit) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.GenDecl:
		if comments := s.cmap[n]; len(comments) != 0 {
			return s.WithComments(comments)
		}
	case ast.Stmt:
		if comments := s.cmap[n]; len(comments) != 0 {
			return s.WithComments(comments)
		}
	case *ast.TypeSpec:
		name := n.Name.Name
		if typ, ok := n.Type.(*ast.StructType); ok {
			if structFieldSelect(name) {
				s.executor(name, s.Comments, typ)
				s.rangeField(typ.Fields)
			}
		}
		return nil
	case *ast.StructType:
		if structFieldSelect("") {
			s.executor("", s.Comments, n)
			s.rangeField(n.Fields)
		}
		return nil
	}
	return s
}
