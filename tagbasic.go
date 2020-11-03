/*
 * Copyright 2020 bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 *
 */

package main

import "go/ast"

type toyVisit struct {
	executor func(name string, n *ast.StructType)
}

func (s *toyVisit) Visit(node ast.Node) ast.Visitor {

	var rangeField func(executor func(name string, n *ast.StructType), fields *ast.FieldList)
	rangeField = func(executor func(name string, n *ast.StructType), fields *ast.FieldList) {
		if fields != nil {
			for _, f := range fields.List {
				if _struct, ok := f.Type.(*ast.StructType); ok {
					executor("", _struct)
					rangeField(executor, _struct.Fields)
				}
			}
		}
	}
	switch n := node.(type) {
	case *ast.TypeSpec:
		name := n.Name.Name
		if typ, ok := n.Type.(*ast.StructType); ok {
			if structFieldSelect(name) {
				s.executor(name, typ)
				rangeField(s.executor, typ.Fields)
			}
		}
		return nil
	case *ast.StructType:
		if structFieldSelect("") {
			s.executor("", n)
			rangeField(s.executor, n.Fields)
		}
		return nil
	}
	return s
}
