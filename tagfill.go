///*
// * Copyright 2020 bigpigeon. All rights reserved.
// * Use of this source code is governed by a MIT style
// * license that can be found in the LICENSE file.
// *
// */
//
package main

import (
	"errors"
	"go/ast"
	"go/token"
	"sort"
	"strings"
)

type tagFillerFields struct {
	fields []*ast.Field
	keySet map[string]struct{}
}

type tagFieldRule func(fieldName string, tagName string) (newTagName string)

type tagFiller struct {
	Err          error
	f            *ast.File
	fs           *token.FileSet
	ruleSet      map[string]tagFieldRule
	needFillList []tagFillerFields
}

func ruleSetClone(rs map[string]tagFieldRule) map[string]tagFieldRule {
	newRs := map[string]tagFieldRule{}
	for k, v := range rs {
		newRs[k] = v
	}
	return newRs
}

func (s *tagFiller) Scan() error {
	ast.Walk(s, s.f)
	return s.Err
}

func (s *tagFiller) Execute() error {
	for _, needFill := range s.needFillList {
		fieldsTagFill(needFill.fields, needFill.keySet, s.ruleSet)
	}
	return nil
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
					s.needFillList = append(s.needFillList, tagFillerFields{n.Fields.List[start:end], keySet})
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
				s.needFillList = append(s.needFillList, tagFillerFields{n.Fields.List[start:], keySet})

			}
		}
	}
	return s
}

func fieldsTagFill(fields []*ast.Field, keySet map[string]struct{}, ruleSet map[string]tagFieldRule) {
	for _, f := range fields {

		if f.Tag != nil {
			rs := ruleSetClone(ruleSet)
			fillMissing := ruleSet["*"]
			delete(rs, "*")
			var appendKeyValues []KeyValue
			quote, keyValues, err := ParseTag(f.Tag.Value)
			if err != nil {
				// must be nil error
				panic(err)
			}
			if fillMissing != nil {
				missingKeySet := keySetClone(keySet)

				for _, kv := range keyValues {
					delete(missingKeySet, kv.Key)
				}
				for k := range rs {
					delete(missingKeySet, k)
				}
				missingKeys := make([]string, 0, len(missingKeySet))
				for k := range missingKeySet {
					missingKeys = append(missingKeys, k)
				}
				sort.Strings(missingKeys)

				for _, k := range missingKeys {
					appendKeyValues = append(appendKeyValues, KeyValue{
						Key:   k,
						quote: quote,
						Value: fillMissing(f.Names[0].Name, ""),
					})
				}

				f.Tag.ValuePos = 0
			}
			missingRuleSet := ruleSetClone(rs)

			for _, kv := range keyValues {
				if rs[kv.Key] != nil {
					kv.Value = rs[kv.Key](f.Names[0].Name, kv.Value)
				}
			}
			for _, kv := range keyValues {
				delete(missingRuleSet, kv.Key)
			}
			for k, rule := range missingRuleSet {
				appendKeyValues = append(appendKeyValues, KeyValue{
					Key:   k,
					quote: quote,
					Value: rule(f.Names[0].Name, ""),
				})
			}
			var keyValueRaw []string
			for _, v := range keyValues {
				keyValueRaw = append(keyValueRaw, v.String())
			}
			for _, v := range appendKeyValues {
				keyValueRaw = append(keyValueRaw, v.String())
			}
			f.Tag.Value = quote + strings.TrimRight(strings.Join(keyValueRaw, " "), " ") + quote
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

func parseFieldRuleSingle(r string) (tagFieldRule, error) {
	r = strings.TrimSpace(r)
	if strings.HasSuffix(r, ")") { // function rule
		bi := strings.Index(r, "(")
		if bi == -1 {
			return nil, errors.New("parse rule failure, invalid rule string")
		}
		subRule, err := parseFieldRuleSingle(r[bi+1 : len(r)-1])
		if err != nil {
			return nil, err
		}
		switch r[:bi] {
		case "upper":
			return func(fieldName string, tagName string) (newTagName string) {
				return strings.ToUpper(subRule(fieldName, tagName))
			}, nil
		case "lower":
			return func(fieldName string, tagName string) (newTagName string) {
				return strings.ToLower(subRule(fieldName, tagName))
			}, nil
		case "hungary":
			return func(fieldName string, tagName string) (newTagName string) {
				return hungaryConvert(subRule(fieldName, tagName))
			}, nil
		case "big_camel":
			return func(fieldName string, tagName string) (newTagName string) {
				return bigCamelConvert(subRule(fieldName, tagName))
			}, nil
		case "lit_camel":
			return func(fieldName string, tagName string) (newTagName string) {
				return litCamelConvert(subRule(fieldName, tagName))
			}, nil
		case "if_not_present":
			return func(fieldName string, tagName string) (newTagName string) {
				if tagName != "" {
					return tagName
				}
				return subRule(fieldName, tagName)
			}, nil
		case "attach_tag_extra":
			return func(fieldName string, tagName string) (newTagName string) {
				commaIndex := strings.Index(tagName, ",")
				if commaIndex != -1 {
					return subRule(fieldName, tagName) + tagName[commaIndex:]
				}
				return subRule(fieldName, tagName)
			}, nil
		default:
			return nil, errors.New("invalid field rule " + r[:bi])
		}
	} else if r == ":field" { // fetch field name
		return func(fieldName string, tagName string) (newTagName string) {
			return fieldName
		}, nil
	} else if r == ":tag" { // fetch field name
		return func(fieldName string, tagName string) (newTagName string) {
			return tagName
		}, nil
	} else { // fetch rule string
		return func(fieldName string, tagName string) (newTagName string) {
			return r
		}, nil
	}
}

func parseFieldRule(s string) (map[string]tagFieldRule, error) {
	rules := map[string]tagFieldRule{}
	for _, cell := range strings.Split(s, "|") {

		keyVal := strings.SplitN(cell, "=", 2)
		if len(keyVal) == 1 {
			rules[keyVal[0]] = func(fieldName string, tagName string) (newTagName string) {
				return ""
			}
			continue
		}
		if len(keyVal) != 2 {
			return nil, errors.New("invalid fill rule (" + cell + ")")
		}

		rule, err := parseFieldRuleSingle(keyVal[1])
		if err != nil {
			return nil, err
		}
		rules[keyVal[0]] = rule
	}
	return rules, nil
}

func newTagFill(f *ast.File, fs *token.FileSet, rule string) (*tagFiller, error) {
	ruleSet, err := parseFieldRule(rule)
	if err != nil {
		return nil, err
	}
	s := &tagFiller{fs: fs, f: f, ruleSet: ruleSet}
	return s, nil
}

func hungaryConvert(name string) string {
	if len(name) == 0 {
		panic("error length name string")
	}
	var convert []byte

	var lowerCount, upperCount uint32
	for i := 0; i < len(name); i++ {
		a := name[i]
		if a >= 'A' && a <= 'Z' {
			if lowerCount >= 1 {
				convert = append(convert, '_', a-'A'+'a')
			} else {
				convert = append(convert, a-'A'+'a')
			}
			upperCount++
			lowerCount = 0
		} else if a >= 'a' && a <= 'z' {
			if upperCount > 1 {
				convert = append(convert, '_', a)
			} else {
				convert = append(convert, a)
			}
			upperCount = 0
			lowerCount++
		} else {
			lowerCount, upperCount = 0, 0
			convert = append(convert, a)
		}
	}
	return string(convert)
}

func bigCamelConvert(name string) string {
	if len(name) == 0 {
		panic("error length name string")
	}
	toBigCamel := false
	var newName []byte
	if name[0] >= 'a' && name[0] <= 'z' {
		newName = append(newName, name[0]-'a'+'A')
		name = name[1:]
	}

	for _, c := range name {
		if c == '_' {
			toBigCamel = true
		} else if toBigCamel {
			if c >= 'a' && c <= 'z' {
				newName = append(newName, byte(c)-'a'+'A')
			}
			toBigCamel = false
		} else {
			newName = append(newName, []byte(string([]rune{c}))...)
		}
	}
	return string(newName)
}

func litCamelConvert(name string) string {
	if len(name) == 0 {
		panic("error length name string")
	}

	toBigCamel := false
	var newName []byte
	if name[0] >= 'A' && name[0] <= 'Z' {
		newName = append(newName, name[0]-'A'+'a')
		name = name[1:]
	}

	for _, c := range name {
		if c == '_' {
			toBigCamel = true
		} else if toBigCamel {
			if c >= 'a' && c <= 'z' {
				newName = append(newName, byte(c)-'a'+'A')
			}
			toBigCamel = false
		} else {
			newName = append(newName, []byte(string([]rune{c}))...)
		}
	}
	return string(newName)
}
