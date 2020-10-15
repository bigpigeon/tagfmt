/*
 * Copyright 2020. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/ast"
	"testing"
)

func TestBigCamelConvert(t *testing.T) {
	assert.Equal(t, bigCamelConvert("id"), "Id")
	assert.Equal(t, bigCamelConvert("bigPigeon"), "BigPigeon")
	assert.Equal(t, bigCamelConvert("big_pigeon"), "BigPigeon")
}

func TestLitCamelConvert(t *testing.T) {
	assert.Equal(t, litCamelConvert("ID"), "iD")
	assert.Equal(t, litCamelConvert("BigPigeon"), "bigPigeon")
	assert.Equal(t, litCamelConvert("big_pigeon"), "bigPigeon")
}

func TestHungaryConvert(t *testing.T) {
	assert.Equal(t, hungaryConvert("UserDetail"), "user_detail")
	assert.Equal(t, hungaryConvert("OneToOne"), "one_to_one")
	assert.Equal(t, hungaryConvert("_UserDetail"), "_user_detail")
	assert.Equal(t, hungaryConvert("userDetail"), "user_detail")
	assert.Equal(t, hungaryConvert("UserDetailID"), "user_detail_id")
	assert.Equal(t, hungaryConvert("NameHTTPtest"), "name_http_test")
	assert.Equal(t, hungaryConvert("IDandValue"), "id_and_value")
	assert.Equal(t, hungaryConvert("toyorm.User.field"), "toyorm.user.field")
}

func TestParseFieldRule(t *testing.T) {
	testFieldArgs := func(name string, oldTag string) *ruleFuncArgs {
		return newRuleArgs(&ast.Field{
			Names: []*ast.Ident{{Name: name}},
		}, oldTag)
	}
	{
		rules, err := parseFieldRule("json=hungary(:field)|yaml=lit_camel(:field)")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", "")), "user_detail")
		assert.Equal(t, rules["yaml"](testFieldArgs("UserDetail", "")), "userDetail")
	}
	{
		rules, err := parseFieldRule("json=or(:tag, hungary(:field))")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", "customUserDetail")), "customUserDetail")
	}

	{
		rules, err := parseFieldRule("json=hungary(:field)+s+:tag_extra")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", ",omitempty")), "user_details,omitempty")
	}

	{
		rules, err := parseFieldRule("json=hungary(:field)+',omitempty'")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", "")), "user_detail,omitempty")
	}
	{
		rules, err := parseFieldRule("json=or(':tag',':field')")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", "user_detail")), "user_detail")
	}
	{
		rules, err := parseFieldRule("json=or(':field', ':tag')")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", "user_detail")), "UserDetail")
	}
	{
		rules, err := parseFieldRule("binding='a|b|c+d,e'")
		require.NoError(t, err)
		assert.Equal(t, rules["binding"](testFieldArgs("UserDetail", "")), "a|b|c+d,e")
	}
}
