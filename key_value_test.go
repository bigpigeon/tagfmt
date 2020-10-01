/*
 * Copyright 2020. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeyValueParse(t *testing.T) {
	s1 := "`json:\"typ1\" yaml:\"typ1\"`"
	s2 := "\"json:\\\"typ2\\\" yaml:\\\"typ2\\\"\""

	quote, kv, err := ParseTag(s1)
	require.NoError(t, err)
	t.Log("quote ", quote, "kv", kv)
	quote, kv, err = ParseTag(string(s2))
	require.NoError(t, err)
	t.Log("quote ", quote, "kv", kv)
}
