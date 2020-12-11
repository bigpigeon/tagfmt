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

type ruleFuncArgs struct {
	Field  *ast.Field
	OldTag string // old tag value
}

func newRuleArgs(f *ast.Field, oldTag string) *ruleFuncArgs {
	return &ruleFuncArgs{
		Field:  f,
		OldTag: oldTag,
	}
}

type tagFieldRule func(field *ruleFuncArgs) (newTagName string)

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
	visit := toyVisit{executor: s.executor}
	return visit.Visit(node)
}

func (s *tagFiller) executor(name string, n *ast.StructType) {
	if n.Fields != nil {
		keySet := map[string]struct{}{}
		var cacheFieldList []*ast.Field
		var preFieldLine int
		for _, field := range n.Fields.List {
			fieldName := getFieldOrTypeName(field)
			if fieldFilter(fieldName) == false {
				continue
			}
			line := s.fs.Position(field.Pos()).Line
			// If there are blank lines or nil field tag in the structure, reset
			if field.Tag == nil || preFieldLine+1 < line {
				s.needFillList = append(s.needFillList, tagFillerFields{cacheFieldList, keySet})
				keySet = map[string]struct{}{}
				cacheFieldList = nil
			}
			preFieldLine = line
			if field.Tag != nil {
				_, keyValues, err := ParseTag(field.Tag.Value)
				if err != nil {
					s.Err = NewAstError(s.fs, field.Tag, err)
					return
				}
				cacheFieldList = append(cacheFieldList, field)
				for _, kv := range keyValues {
					keySet[kv.Key] = struct{}{}
				}
			}
		}
		if cacheFieldList != nil {
			s.needFillList = append(s.needFillList, tagFillerFields{cacheFieldList, keySet})
		}
	}
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

				for _, k := range missingKeys {
					appendKeyValues = append(appendKeyValues, KeyValue{
						Key:   k,
						quote: quote,
						Value: fillMissing(newRuleArgs(f, "")),
					})
				}

				f.Tag.ValuePos = 0
			}
			missingRuleSet := ruleSetClone(rs)

			for i, kv := range keyValues {
				if rs[kv.Key] != nil {
					keyValues[i].Value = rs[kv.Key](newRuleArgs(f, kv.Value))
				}
			}
			for _, kv := range keyValues {
				delete(missingRuleSet, kv.Key)
			}

			for k, rule := range missingRuleSet {
				appendKeyValues = append(appendKeyValues, KeyValue{
					Key:   k,
					quote: quote,
					Value: rule(newRuleArgs(f, "")),
				})
			}
			sort.Slice(appendKeyValues, func(i, j int) bool {
				return appendKeyValues[i].Key < appendKeyValues[j].Key
			})
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
	for i := 0; i < len(r); i++ {
		switch _c := r[i]; _c {
		case '(':
			c++
		case ')':
			if c == 0 {
				return i
			}
			c--
		case '\'', '"':
			i++
			ni := findNextQuote(r, i, _c)
			if ni == -1 {
				return -1
			}
			i = ni
		}

	}
	return -1
}

func splitPlusSign(r string) ([]string, error) {
	r = strings.TrimSpace(r)
	var rule []string
	s, e := 0, 0
	for ; e < len(r); e++ {
		switch c := r[e]; c {
		case '+':
			rule = append(rule, r[s:e])
			s = e + 1
		case '(':
			e++
			i := findRightBracket(r[e:])
			if i == -1 {
				return nil, ErrUnclosedBracket
			}
			e += i
		case '\'', '"':
			e++
			ni := findNextQuote(r, e, c)
			if ni == -1 {
				return nil, ErrUnclosedQuote
			}
			e = ni
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
		argsStr := r[bi+1 : len(r)-1]
		switch r[:bi] {
		case "upper":
			subRuleList, err := parseFieldMultiRule(argsStr, 1)
			if err != nil {
				return nil, err
			}
			return func(args *ruleFuncArgs) (newTagName string) {
				return strings.ToUpper(subRuleList[0](args))
			}, nil
		case "lower":
			subRuleList, err := parseFieldMultiRule(argsStr, 1)
			if err != nil {
				return nil, err
			}
			return func(args *ruleFuncArgs) (newTagName string) {
				return strings.ToLower(subRuleList[0](args))
			}, nil
		case "snake":
			subRuleList, err := parseFieldMultiRule(argsStr, 1)
			if err != nil {
				return nil, err
			}
			return func(args *ruleFuncArgs) (newTagName string) {
				return snakeConvert(subRuleList[0](args))
			}, nil
		case "upper_camel":
			subRuleList, err := parseFieldMultiRule(argsStr, 1)
			if err != nil {
				return nil, err
			}
			return func(args *ruleFuncArgs) (newTagName string) {
				return upperCamelConvert(subRuleList[0](args))
			}, nil
		case "lower_camel":
			subRuleList, err := parseFieldMultiRule(argsStr, 1)
			if err != nil {
				return nil, err
			}
			return func(args *ruleFuncArgs) (newTagName string) {
				return lowerCamelConvert(subRuleList[0](args))
			}, nil
		case "or":
			subRuleList, err := parseFieldMultiRule(argsStr, 2)
			if err != nil {
				return nil, err
			}
			return func(args *ruleFuncArgs) (newTagName string) {
				val1 := subRuleList[0](args)
				if val1 != "" {
					return val1
				}
				return subRuleList[1](args)
			}, nil
		default:
			return nil, errors.New("invalid field rule " + r[:bi])
		}
	} else {
		if len(r) > 0 && (r[0] == '\'' || r[0] == '"') {
			r = strings.Trim(r, string(r[0]))
		}
		if r == ":field" { // fetch field name
			return func(args *ruleFuncArgs) (newTagName string) {
				return getFieldName(args.Field)
			}, nil
		} else if r == ":tag" { // fetch field name
			return func(args *ruleFuncArgs) (newTagName string) {
				return args.OldTag
			}, nil
		} else if r == ":tag_basic" { // fetch tag before ','
			return func(args *ruleFuncArgs) (newTagName string) {
				if comma := strings.Index(args.OldTag, ","); comma != -1 {
					return args.OldTag[:comma]
				}
				return args.OldTag
			}, nil
		} else if r == ":tag_extra" { // fetch tag after ','
			return func(args *ruleFuncArgs) (newTagName string) {
				if comma := strings.Index(args.OldTag, ","); comma != -1 {
					return args.OldTag[comma:]
				}
				return ""
			}, nil
		} else {
			return func(args *ruleFuncArgs) (newTagName string) {
				return r
			}, nil
		}
	}
}

func findNextQuote(s string, i int, quote byte) int {
	for j := i; j < len(s); j++ {
		switch o := s[j]; o {
		case '\\':
			j++
		case quote:
			return j
		}
	}
	return -1
}

func splitWithoutQuote(s string, key byte) ([]string, error) {
	var sub []string
	pre := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if s[i] == '"' || s[i] == '\'' {
			nextQuote := findNextQuote(s, i+1, c)
			if nextQuote == -1 {
				return nil, ErrUnclosedQuote
			}
			i = nextQuote
		} else if s[i] == key {
			sub = append(sub, s[pre:i])
			pre = i + 1

		}
	}
	if pre != len(s) {
		sub = append(sub, s[pre:])
	}
	return sub, nil
}

// parse multiple rule, split with ',',
// r: is the rule string
// argsNum: args number limit, return error if args not equal to the argsNum
// e.g: parseFieldMultiRule(":tag, My+',omitempty'", 2) => will get two tagFieldRule
func parseFieldMultiRule(r string, argsNum int) ([]tagFieldRule, error) {
	r = strings.TrimSpace(r)
	rSplitComma, err := splitWithoutQuote(r, ',')
	if err != nil {
		return nil, err
	}
	if len(rSplitComma) != argsNum {
		return nil, errors.New("args number wrong")
	}
	var ruleList []tagFieldRule
	for _, rule := range rSplitComma {
		ruleStrList, err := splitPlusSign(rule)
		if err != nil {
			return nil, err
		}
		var subRules []tagFieldRule
		for _, r := range ruleStrList {
			single, err := parseFieldRuleSingle(r)
			if err != nil {
				return nil, err
			}
			subRules = append(subRules, single)
		}
		ruleList = append(ruleList, func(info *ruleFuncArgs) (newTagName string) {
			s := ""
			for _, rule := range subRules {
				s += rule(info)
			}
			return s
		})
	}
	return ruleList, nil
}

// parse with '+' rule
func parseFieldRulePlus(r string) (tagFieldRule, error) {
	ruleStrList, err := splitPlusSign(r)
	if err != nil {
		return nil, err
	}
	var subRules []tagFieldRule
	for _, r := range ruleStrList {
		single, err := parseFieldRuleSingle(r)
		if err != nil {
			return nil, err
		}
		subRules = append(subRules, single)
	}
	return func(info *ruleFuncArgs) (newTagName string) {
		s := ""
		for _, rule := range subRules {
			s += rule(info)
		}
		return s
	}, nil
}

func parseFieldRule(s string) (map[string]tagFieldRule, error) {
	rules := map[string]tagFieldRule{}
	var err error
	ruleList, err := splitWithoutQuote(s, '|')
	if err != nil {
		return nil, err
	}
	for _, cell := range ruleList {
		keyVal := strings.SplitN(cell, "=", 2)
		// if value is nil ,use key hold rule
		if len(keyVal) == 1 {
			rules[keyVal[0]], err = parseFieldRulePlus("")
			if err != nil { // error must be not nil
				panic(err)
			}
			rules[keyVal[0]] = func(info *ruleFuncArgs) (newTagName string) {
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

func snakeConvert(name string) string {
	if len(name) == 0 {
		return ""
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

func upperCamelConvert(name string) string {
	if len(name) == 0 {
		return ""
	}
	toUpperCamel := false
	var newName []byte
	if name[0] >= 'a' && name[0] <= 'z' {
		newName = append(newName, name[0]-'a'+'A')
		name = name[1:]
	}

	for _, c := range name {
		if c == '_' {
			toUpperCamel = true
		} else if toUpperCamel {
			if c >= 'a' && c <= 'z' {
				newName = append(newName, byte(c)-'a'+'A')
			} else {
				newName = append(newName, []byte(string([]rune{c}))...)
			}
			toUpperCamel = false
		} else {
			newName = append(newName, []byte(string([]rune{c}))...)
		}
	}
	return string(newName)
}

func lowerCamelConvert(name string) string {
	if len(name) == 0 {
		return ""
	}

	toUpperCamel := false
	var newName []byte
	if name[0] >= 'A' && name[0] <= 'Z' {
		newName = append(newName, name[0]-'A'+'a')
		name = name[1:]
	}

	for _, c := range name {
		if c == '_' {
			toUpperCamel = true
		} else if toUpperCamel {
			if c >= 'a' && c <= 'z' {
				newName = append(newName, byte(c)-'a'+'A')
			} else {
				newName = append(newName, []byte(string([]rune{c}))...)
			}
			toUpperCamel = false
		} else {
			newName = append(newName, []byte(string([]rune{c}))...)
		}
	}
	return string(newName)
}
