/*
 * Copyright 2020. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package main

type KeyValue struct {
	Key   string
	quote string
	Value string
}

func (kv KeyValue) String() string {
	if kv.quote == "`" {
		return kv.Key + `:"` + kv.Value + `"`
	} else if kv.quote == "\"" {
		return kv.Key + `:\\"` + kv.Value + `:\\"`
	} else {
		panic("invalid quote " + kv.quote)
	}
}

// ParseTag returns all tag keys and tags key:"Value" list
func ParseTag(tag string) (quote string, keyValues []KeyValue, err error) {
	if len(tag) < 2 {
		err = ErrInvalidTag
		return
	}

	quote = tag[:1]
	tag = tag[1 : len(tag)-1]

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		name := string(tag[:i])
		if i == 0 || i >= len(tag) || tag[i] != ':' {
			return "", nil, ErrInvalidTag
		}
		i++
		var quoteLen int
		if i >= len(tag) || tag[i] != '"' {
			if tag[i] != '\\' || i+1 >= len(tag) || tag[i+1] != '"' {
				return "", nil, ErrInvalidTag
			}
			tag = tag[i:]
			i = 2
			quoteLen = 2
		} else {
			tag = tag[i:]
			i = 1
			quoteLen = 1
		}

		// Scan quoted string to find value.

		for i < len(tag) && tag[i] != '"' {
			i++
		}
		if quoteLen == 2 && tag[i-1] != '\\' {
			return "", nil, ErrInvalidTag
		}
		if i >= len(tag) {
			return "", nil, ErrInvalidTag
		}
		value := string(tag[quoteLen : i-quoteLen+1])

		keyValues = append(keyValues, KeyValue{
			Key:   name,
			Value: value,
			quote: quote,
		})

		tag = tag[i+1:]
	}
	return
}
