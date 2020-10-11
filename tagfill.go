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
				if fieldFilter(field.Names[0].Name) == false {
					continue
				}
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

			for i, kv := range keyValues {
				if rs[kv.Key] != nil {
					keyValues[i].Value = rs[kv.Key](f.Names[0].Name, kv.Value)
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

func findRightBracket(r string) int {
	c := 0 // extra left bracket count
	for e := 0; e < len(r); e++ {
		if r[e] == '(' {
			c++
		} else if r[e] == ')' {
			if c == 0 {
				return e
			}
			c--
		}
	}
	return -1
}

func parsePlusSign(r string) ([]string, error) {
	var rule []string
	s, e := 0, 0
	for ; e < len(r); e++ {
		switch r[e] {
		case '+':
			rule = append(rule, r[s:e])
			s = e + 1
		case '(':
			e++
			i := findRightBracket(r[e:])
			if i == -1 {
				return nil, errors.New("invalid rule:" + r)
			}
			e += i
		}
	}
	if s != e {
		rule = append(rule, r[s:e])
	}
	return rule, nil
}

func parseFieldRuleSingle(r string) (tagFieldRule, error) {
	if strings.HasSuffix(r, ")") { // function rule
		bi := strings.Index(r, "(")
		if bi == -1 {
			return nil, errors.New("parse rule failure, invalid rule string")
		}
		subRule, err := parseFieldRulePlus(r[bi+1 : len(r)-1])
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
	} else if r == ":tag_basic" { // fetch tag before ','
		return func(fieldName string, tagName string) (newTagName string) {
			if comma := strings.Index(tagName, ","); comma != -1 {
				return tagName[:comma]
			}
			return tagName
		}, nil
	} else if r == ":tag_extra" { // fetch tag after ','
		return func(fieldName string, tagName string) (newTagName string) {
			if comma := strings.Index(tagName, ","); comma != -1 {
				return tagName[comma:]
			}
			return ""
		}, nil
	} else {
		return func(fieldName string, tagName string) (newTagName string) {
			return r
		}, nil
	}
}

// parse with '+' rule
func parseFieldRulePlus(r string) (tagFieldRule, error) {
	r = strings.TrimSpace(r)
	ruleStrList, err := parsePlusSign(r)
	if err != nil {
		return nil, err
	}
	var rules []tagFieldRule
	for _, r := range ruleStrList {
		single, err := parseFieldRuleSingle(r)
		if err != nil {
			return nil, err
		}
		rules = append(rules, single)
	}
	return func(fieldName string, tagName string) (newTagName string) {
		s := ""
		for _, rule := range rules {
			s += rule(fieldName, tagName)
		}
		return s
	}, nil

}

func parseFieldRule(s string) (map[string]tagFieldRule, error) {
	rules := map[string]tagFieldRule{}
	var err error
	for _, cell := range strings.Split(s, "|") {

		keyVal := strings.SplitN(cell, "=", 2)
		// if value is nil ,use key hold rule
		if len(keyVal) == 1 {
			rules[keyVal[0]], err = parseFieldRulePlus("")
			if err != nil { // error must be not nil
				panic(err)
			}
			rules[keyVal[0]] = func(fieldName string, tagName string) (newTagName string) {
				return ""
			}
			continue
		}
		if len(keyVal) != 2 {
			return nil, errors.New("invalid fill rule (" + cell + ")")
		}

		rule, err := parseFieldRulePlus(keyVal[1])
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
			} else {
				newName = append(newName, []byte(string([]rune{c}))...)
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
			} else {
				newName = append(newName, []byte(string([]rune{c}))...)
			}
			toBigCamel = false
		} else {
			newName = append(newName, []byte(string([]rune{c}))...)
		}
	}
	return string(newName)
}
