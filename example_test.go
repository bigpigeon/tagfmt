/*
 * Copyright 2020 bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 *
 */

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func ExampleAlignWrite() {
	resetFlags()
	bakData, err := ioutil.ReadFile("exampledata/api.go")
	if err != nil {
		panic(err)
	}

	bakName, err := backupFile("exampledata/api.go.bak", bakData, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := os.Remove("exampledata/api.go")
		if err != nil {
			panic(err)
		}
		err = os.Rename(bakName, "exampledata/api.go")
		if err != nil {
			panic(err)
		}
	}()
	os.Args = strings.Split("tagfmt -w exampledata/api.go", " ")
	data, err := ioutil.ReadFile("exampledata/api.go")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	gofmtMain()
	//output:
	//package exampledata
	//
	//type OrderDetail struct {
	//	ID       string   `json:"id" yaml:"id"`
	//	UserName string   `json:"user_name" yaml:"user_name"`
	//	OrderID  string   `json:"order_id" yaml:"order_id"`
	//	Callback string   `json:"callback" yaml:"callback"`
	//	Address  []string `json:"address" yaml:"address"`
	//}
}

func ExampleAlign() {
	resetFlags()
	os.Args = strings.Split("tagfmt exampledata/", " ")
	gofmtMain()
	// output:
	//package exampledata
	//
	//type OrderDetail struct {
	//	ID       string   `json:"id"        yaml:"id"`
	//	UserName string   `json:"user_name" yaml:"user_name"`
	//	OrderID  string   `json:"order_id"  yaml:"order_id"`
	//	Callback string   `json:"callback"  yaml:"callback"`
	//	Address  []string `json:"address"   yaml:"address"`
	//}
}
