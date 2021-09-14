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

func TestUpperCamelConvert(t *testing.T) {
	assert.Equal(t, upperCamelConvert("id"), "Id")
	assert.Equal(t, upperCamelConvert("bigPigeon"), "BigPigeon")
	assert.Equal(t, upperCamelConvert("big_pigeon"), "BigPigeon")
}

func TestLowerCamelConvert(t *testing.T) {
	assert.Equal(t, lowerCamelConvert("ID"), "iD")
	assert.Equal(t, lowerCamelConvert("BigPigeon"), "bigPigeon")
	assert.Equal(t, lowerCamelConvert("big_pigeon"), "bigPigeon")
}

func TestSnakeConvert(t *testing.T) {
	assert.Equal(t, snakeConvert("UserDetail"), "user_detail")
	assert.Equal(t, snakeConvert("OneToOne"), "one_to_one")
	assert.Equal(t, snakeConvert("_UserDetail"), "_user_detail")
	assert.Equal(t, snakeConvert("userDetail"), "user_detail")
	assert.Equal(t, snakeConvert("UserDetailID"), "user_detail_id")
	assert.Equal(t, snakeConvert("NameHTTPtest"), "name_http_test")
	assert.Equal(t, snakeConvert("IDandValue"), "id_and_value")
	assert.Equal(t, snakeConvert("toyorm.User.field"), "toyorm.user.field")
}

func TestNewSnakeConvert(t *testing.T) {
	assert.Equal(t, newSnakeConvert("ID"), "id")
	assert.Equal(t, newSnakeConvert("UserDetail"), "user_detail")
	assert.Equal(t, newSnakeConvert("OneToOne"), "one_to_one")
	assert.Equal(t, newSnakeConvert("_UserDetail"), "_user_detail")
	assert.Equal(t, newSnakeConvert("userDetail"), "user_detail")
	assert.Equal(t, newSnakeConvert("UserDetailID"), "user_detail_id")
	assert.Equal(t, newSnakeConvert("NameHTTPTest"), "name_http_test")
	assert.Equal(t, newSnakeConvert("IDAndValue"), "id_and_value")
	assert.Equal(t, newSnakeConvert("toyorm.User.field"), "toyorm.user.field")
	assert.Equal(t, newSnakeConvert("USER"), "user")
	assert.Equal(t, newSnakeConvert("User98Test"), "user98test")
}

func TestParseFieldRule(t *testing.T) {
	testFieldArgs := func(name string, oldTag string) *ruleFuncArgs {
		return newRuleArgs(&ast.Field{
			Names: []*ast.Ident{{Name: name}},
		}, oldTag)
	}
	{
		rules, err := parseFieldRule("json=snake(:field)|yaml=lower_camel(:field)")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", "")), "user_detail")
		assert.Equal(t, rules["yaml"](testFieldArgs("UserDetail", "")), "userDetail")
	}
	{
		rules, err := parseFieldRule("json=or(:tag, snake(:field))")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", "customUserDetail")), "customUserDetail")
	}

	{
		rules, err := parseFieldRule("json=snake(:field)+s+:tag_extra")
		require.NoError(t, err)
		assert.Equal(t, rules["json"](testFieldArgs("UserDetail", ",omitempty")), "user_details,omitempty")
	}

	{
		rules, err := parseFieldRule("json=snake(:field)+',omitempty'")
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
