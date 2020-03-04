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

// BuildInTypes defines the correspondence between Go, TypeScript, C, Java
// languages and data types in XSD.
// https://www.w3.org/TR/xmlschema-2/#datatype
var BuildInTypes = map[string][]string{
	"anyType":            {"string", "string", "char", "String"},
	"ENTITIES":           {"[]string", "Array<string>", "char[]", "List<String>"},
	"ENTITY":             {"string", "string", "char", "String"},
	"ID":                 {"string", "string", "char", "String"},
	"IDREF":              {"string", "string", "char", "String"},
	"IDREFS":             {"[]string", "Array<string>", "char[]", "List<String>"},
	"NCName":             {"string", "string", "char", "String"},
	"NMTOKEN":            {"string", "string", "char", "String"},
	"NMTOKENS":           {"[]string", "Array<string>", "char[]", "List<String>"},
	"NOTATION":           {"[]string", "Array<string>", "char[]", "List<String>"},
	"Name":               {"string", "string", "char", "String"},
	"QName":              {"xml.Name", "any", "char", "String"},
	"anyURI":             {"string", "string", "char", "QName"},
	"base64Binary":       {"[]byte", "Array<any>", "char[]", "List<Byte>"},
	"boolean":            {"bool", "boolean", "bool", "Boolean"},
	"byte":               {"byte", "any", "char[]", "Byte"},
	"date":               {"time.Time", "string", "char", "Byte"},
	"dateTime":           {"time.Time", "string", "char", "Byte"},
	"decimal":            {"float64", "number", "float", "Float"},
	"double":             {"float64", "number", "float", "Float"},
	"duration":           {"string", "string", "char", "String"},
	"float":              {"float32", "number", "float", "Float"},
	"gDay":               {"time.Time", "string", "char", "String"},
	"gMonth":             {"time.Time", "string", "char", "String"},
	"gMonthDay":          {"time.Time", "string", "char", "String"},
	"gYear":              {"time.Time", "string", "char", "String"},
	"gYearMonth":         {"time.Time", "string", "char", "String"},
	"hexBinary":          {"[]byte", "Array<any>", "char[]", "List<Byte>"},
	"int":                {"int", "number", "int", "Integer"},
	"integer":            {"int", "number", "int", "Integer"},
	"language":           {"string", "string", "char", "String"},
	"long":               {"int64", "number", "int", "Long"},
	"negativeInteger":    {"int", "number", "int", "Integer"},
	"nonNegativeInteger": {"int", "number", "int", "Integer"},
	"normalizedString":   {"string", "string", "char", "String"},
	"nonPositiveInteger": {"int", "number", "int", "Integer"},
	"positiveInteger":    {"int", "number", "int", "Integer"},
	"short":              {"int", "number", "int", "Integer"},
	"string":             {"string", "string", "char", "String"},
	"time":               {"time.Time", "string", "char", "String"},
	"token":              {"string", "string", "char", "String"},
	"unsignedByte":       {"byte", "any", "char", "Byte"},
	"unsignedInt":        {"uint", "number", "unsigned int", "Integer"},
	"unsignedLong":       {"uint64", "number", "unsigned int", "Long"},
	"unsignedShort":      {"uint", "number", "unsigned int", "Short"},
	"xml:lang":           {"string", "string", "char", "String"},
	"xml:space":          {"string", "string", "char", "String"},
	"xml:base":           {"string", "string", "char", "String"},
	"xml:id":             {"string", "string", "char", "String"},
}

func getBuildInTypeByLang(value, lang string) (buildType string, ok bool) {
	var supportLang = map[string]int{
		"Go":         0,
		"TypeScript": 1,
		"C":          2,
		"Java":       3,
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
