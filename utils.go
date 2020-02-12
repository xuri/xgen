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
var buildInTypes = map[string]string{
	"anyType":            "string",
	"ENTITIES":           "[]string",
	"ENTITY":             "string",
	"ID":                 "string",
	"IDREF":              "string",
	"IDREFS":             "[]string",
	"NCName":             "string",
	"NMTOKEN":            "string",
	"NMTOKENS":           "[]string",
	"NOTATION":           "[]string",
	"Name":               "string",
	"QName":              "xml.Name",
	"anyURI":             "string",
	"base64Binary":       "[]byte",
	"boolean":            "bool",
	"byte":               "byte",
	"date":               "time.Time",
	"dateTime":           "time.Time",
	"decimal":            "float64",
	"double":             "float64",
	"duration":           "string",
	"float":              "float32",
	"gDay":               "time.Time",
	"gMonth":             "time.Time",
	"gMonthDay":          "time.Time",
	"gYear":              "time.Time",
	"gYearMonth":         "time.Time",
	"hexBinary":          "[]byte",
	"int":                "int",
	"integer":            "int",
	"language":           "string",
	"long":               "int64",
	"negativeInteger":    "int",
	"nonNegativeInteger": "int",
	"normalizedString":   "string",
	"nonPositiveInteger": "int",
	"positiveInteger":    "int",
	"short":              "int",
	"string":             "string",
	"time":               "time.Time",
	"token":              "string",
	"unsignedByte":       "byte",
	"unsignedInt":        "uint",
	"unsignedLong":       "uint64",
	"unsignedShort":      "uint",
	"xml:lang":           "string",
	"xml:space":          "string",
	"xml:base":           "string",
	"xml:id":             "string",
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
