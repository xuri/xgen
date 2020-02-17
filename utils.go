// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// GetFileList get a list of file by given path.
func GetFileList(path string) (files []string, err error) {
	var fi os.FileInfo
	fi, err = os.Stat(path)
	if err != nil {
		return
	}
	if fi.IsDir() {
		err = filepath.Walk(path, func(fp string, info os.FileInfo, err error) error {
			files = append(files, fp)
			return nil
		})
		if err != nil {
			return
		}
	}
	files = append(files, path)
	return
}

// PrepareOutputDir provide a method to create the output directory by given
// path.
func PrepareOutputDir(path string) error {
	if path == "" {
		return nil
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

// https://www.w3.org/TR/xmlschema-2/#datatype
// XSD, Go, TypeScript
var BuildInTypes = map[string][]string{
	"anyType":            {"string", "string"},
	"ENTITIES":           {"[]string", "Array<string>"},
	"ENTITY":             {"string", "string"},
	"ID":                 {"string", "string"},
	"IDREF":              {"string", "string"},
	"IDREFS":             {"[]string", "Array<string>"},
	"NCName":             {"string", "string"},
	"NMTOKEN":            {"string", "string"},
	"NMTOKENS":           {"[]string", "Array<string>"},
	"NOTATION":           {"[]string", "Array<string>"},
	"Name":               {"string", "string"},
	"QName":              {"xml.Name", "any"},
	"anyURI":             {"string", "string"},
	"base64Binary":       {"[]byte", "Array<any>"},
	"boolean":            {"bool", "boolean"},
	"byte":               {"byte", "any"},
	"date":               {"time.Time", "string"},
	"dateTime":           {"time.Time", "string"},
	"decimal":            {"float64", "number"},
	"double":             {"float64", "number"},
	"duration":           {"string", "string"},
	"float":              {"float32", "number"},
	"gDay":               {"time.Time", "string"},
	"gMonth":             {"time.Time", "string"},
	"gMonthDay":          {"time.Time", "string"},
	"gYear":              {"time.Time", "string"},
	"gYearMonth":         {"time.Time", "string"},
	"hexBinary":          {"[]byte", "Array<any>"},
	"int":                {"int", "number"},
	"integer":            {"int", "number"},
	"language":           {"string", "string"},
	"long":               {"int64", "number"},
	"negativeInteger":    {"int", "number"},
	"nonNegativeInteger": {"int", "number"},
	"normalizedString":   {"string", "string"},
	"nonPositiveInteger": {"int", "number"},
	"positiveInteger":    {"int", "number"},
	"short":              {"int", "number"},
	"string":             {"string", "string"},
	"time":               {"time.Time", "string"},
	"token":              {"string", "string"},
	"unsignedByte":       {"byte", "any"},
	"unsignedInt":        {"uint", "number"},
	"unsignedLong":       {"uint64", "number"},
	"unsignedShort":      {"uint", "number"},
	"xml:lang":           {"string", "string"},
	"xml:space":          {"string", "string"},
	"xml:base":           {"string", "string"},
	"xml:id":             {"string", "string"},
}

func getBuildInTypeByLang(value, lang string) (buildType string, ok bool) {
	var supportLang = map[string]int{
		"Go":         0,
		"TypeScript": 1,
	}
	var buildInTypes []string
	if buildInTypes, ok = BuildInTypes[value]; !ok {
		return
	}
	buildType = buildInTypes[supportLang[lang]]
	return
}
func getBasefromSimpleType(name string, XSDSchema []interface{}) string {
	for _, ele := range XSDSchema {
		switch v := ele.(type) {
		case *SimpleType:
			if !v.List && !v.Union && v.Name == name {
				return v.Base
			}
		case *Attribute:
			if v.Name == name {
				return v.Type
			}
		case *Element:
			if v.Name == name {
				return v.Type
			}
		}
	}
	return name
}

func getNSPrefix(str string) (ns string) {
	split := strings.Split(str, ":")
	if len(split) == 2 {
		ns = split[0]
		return
	}
	return
}

func trimNSPrefix(str string) (name string) {
	split := strings.Split(str, ":")
	if len(split) == 2 {
		name = split[1]
		return
	}
	name = str
	return
}

// MakeFirstUpperCase make the first letter of a string uppercase.
func MakeFirstUpperCase(s string) string {

	if len(s) < 2 {
		return strings.ToUpper(s)
	}

	bts := []byte(s)

	lc := bytes.ToUpper([]byte{bts[0]})
	rest := bts[1:]

	return string(bytes.Join([][]byte{lc, rest}, nil))
}

// callFuncByName calls the only error return function with reflect by given
// receiver, name and parameters.
func callFuncByName(receiver interface{}, name string, params []reflect.Value) (err error) {
	function := reflect.ValueOf(receiver).MethodByName(name)
	if function.IsValid() {
		rt := function.Call(params)
		if !rt[0].IsNil() {
			err = rt[0].Interface().(error)
			return
		}
	}
	return
}
