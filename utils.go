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

// BuildInTypes defines the correspondence betweenGo, TypeScript, C languages
// and data types in XSD. https://www.w3.org/TR/xmlschema-2/#datatype
var BuildInTypes = map[string][]string{
	"anyType":            {"string", "string", "char"},
	"ENTITIES":           {"[]string", "Array<string>", "char[]"},
	"ENTITY":             {"string", "string", "char"},
	"ID":                 {"string", "string", "char"},
	"IDREF":              {"string", "string", "char"},
	"IDREFS":             {"[]string", "Array<string>", "char[]"},
	"NCName":             {"string", "string", "char"},
	"NMTOKEN":            {"string", "string", "char"},
	"NMTOKENS":           {"[]string", "Array<string>", "char[]"},
	"NOTATION":           {"[]string", "Array<string>", "char[]"},
	"Name":               {"string", "string", "char"},
	"QName":              {"xml.Name", "any", "char"},
	"anyURI":             {"string", "string", "char"},
	"base64Binary":       {"[]byte", "Array<any>", "char[]"},
	"boolean":            {"bool", "boolean", "bool"},
	"byte":               {"byte", "any", "char[]"},
	"date":               {"time.Time", "string", "char"},
	"dateTime":           {"time.Time", "string", "char"},
	"decimal":            {"float64", "number", "float"},
	"double":             {"float64", "number", "float"},
	"duration":           {"string", "string", "char"},
	"float":              {"float32", "number", "float"},
	"gDay":               {"time.Time", "string", "char"},
	"gMonth":             {"time.Time", "string", "char"},
	"gMonthDay":          {"time.Time", "string", "char"},
	"gYear":              {"time.Time", "string", "char"},
	"gYearMonth":         {"time.Time", "string", "char"},
	"hexBinary":          {"[]byte", "Array<any>", "char[]"},
	"int":                {"int", "number", "int"},
	"integer":            {"int", "number", "int"},
	"language":           {"string", "string", "char"},
	"long":               {"int64", "number", "int"},
	"negativeInteger":    {"int", "number", "int"},
	"nonNegativeInteger": {"int", "number", "int"},
	"normalizedString":   {"string", "string", "char"},
	"nonPositiveInteger": {"int", "number", "int"},
	"positiveInteger":    {"int", "number", "int"},
	"short":              {"int", "number", "int"},
	"string":             {"string", "string", "char"},
	"time":               {"time.Time", "string", "char"},
	"token":              {"string", "string", "char"},
	"unsignedByte":       {"byte", "any", "char"},
	"unsignedInt":        {"uint", "number", "unsigned int"},
	"unsignedLong":       {"uint64", "number", "unsigned int"},
	"unsignedShort":      {"uint", "number", "unsigned int"},
	"xml:lang":           {"string", "string", "char"},
	"xml:space":          {"string", "string", "char"},
	"xml:base":           {"string", "string", "char"},
	"xml:id":             {"string", "string", "char"},
}

func getBuildInTypeByLang(value, lang string) (buildType string, ok bool) {
	var supportLang = map[string]int{
		"Go":         0,
		"TypeScript": 1,
		"C":          2,
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
