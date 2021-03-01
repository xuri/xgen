// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/net/html/charset"
)

// Options holds user-defined overrides and runtime data that are used when
// parsing from an XSD document.
type Options struct {
	FilePath            string
	FileDir             string
	InputDir            string
	OutputDir           string
	Extract             bool
	Lang                string
	Package             string
	IncludeMap          map[string]bool
	LocalNameNSMap      map[string]string
	NSSchemaLocationMap map[string]string
	ParseFileList       map[string]bool
	ParseFileMap        map[string][]interface{}
	ProtoTree           []interface{}
	RemoteSchema        map[string][]byte

	InElement        string
	CurrentEle       string
	InGroup          int
	InUnion          bool
	InAttributeGroup bool

	SimpleType     *Stack
	ComplexType    *Stack
	Element        *Stack
	Attribute      *Stack
	Group          *Stack
	AttributeGroup *Stack
}

// NewParser creates a new parser options for the Parse. Useful for XML schema
// parsing.
func NewParser(options *Options) *Options {
	return options
}

// Parse reads XML documents and return proto tree for every element in the
// documents by given options. If value of the properity extract is false,
// parse will fetch schema used in <import> or <include> statements.
func (opt *Options) Parse() (err error) {
	opt.FileDir = filepath.Dir(opt.FilePath)
	var fi os.FileInfo
	fi, err = os.Stat(opt.FilePath)
	if err != nil {
		return
	}
	if fi.IsDir() {
		return
	}
	var xmlFile *os.File
	xmlFile, err = os.Open(opt.FilePath)
	if err != nil {
		return
	}
	defer xmlFile.Close()

	if !opt.Extract {
		opt.ParseFileList[opt.FilePath] = true
		opt.ParseFileMap[opt.FilePath] = opt.ProtoTree
	}
	opt.ProtoTree = make([]interface{}, 0)

	opt.InElement = ""
	opt.CurrentEle = ""
	opt.InGroup = 0
	opt.InUnion = false
	opt.InAttributeGroup = false

	opt.SimpleType = NewStack()
	opt.ComplexType = NewStack()
	opt.Element = NewStack()
	opt.Attribute = NewStack()
	opt.Group = NewStack()
	opt.AttributeGroup = NewStack()

	decoder := xml.NewDecoder(xmlFile)
	decoder.CharsetReader = charset.NewReaderLabel
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		switch element := token.(type) {
		case xml.StartElement:

			opt.InElement = element.Name.Local
			funcName := fmt.Sprintf("On%s", MakeFirstUpperCase(opt.InElement))
			if err = callFuncByName(opt, funcName, []reflect.Value{reflect.ValueOf(element), reflect.ValueOf(opt.ProtoTree)}); err != nil {
				return
			}

		case xml.EndElement:
			funcName := fmt.Sprintf("End%s", MakeFirstUpperCase(element.Name.Local))
			if err = callFuncByName(opt, funcName, []reflect.Value{reflect.ValueOf(element), reflect.ValueOf(opt.ProtoTree)}); err != nil {
				return
			}
		case xml.CharData:
			if err = opt.OnCharData(string(element), opt.ProtoTree); err != nil {
				return
			}
		default:
		}

	}

	if !opt.Extract {
		opt.ParseFileList[opt.FilePath] = true
		opt.ParseFileMap[opt.FilePath] = opt.ProtoTree
		path := filepath.Join(opt.OutputDir, strings.TrimPrefix(opt.FilePath, opt.InputDir))
		if err := PrepareOutputDir(filepath.Dir(path)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		generator := &CodeGenerator{
			Lang:      opt.Lang,
			Package:   opt.Package,
			File:      path,
			ProtoTree: opt.ProtoTree,
			StructAST: map[string]string{},
		}
		funcName := fmt.Sprintf("Gen%s", MakeFirstUpperCase(opt.Lang))
		if err = callFuncByName(generator, funcName, []reflect.Value{}); err != nil {
			return
		}
	}
	return
}

// GetValueType convert XSD schema value type to the build-in type for the
// given value and proto tree.
func (opt *Options) GetValueType(value string, XSDSchema []interface{}) (valueType string, err error) {
	if buildType, ok := getBuildInTypeByLang(trimNSPrefix(value), opt.Lang); ok {
		valueType = buildType
		return
	}
	valueType = getBasefromSimpleType(trimNSPrefix(value), XSDSchema)
	if valueType != trimNSPrefix(value) && valueType != "" {
		return
	}
	if opt.Extract {
		return
	}
	schemaLocation := opt.NSSchemaLocationMap[opt.parseNS(value)]
	if isValidURL(schemaLocation) {
		return
	}
	xsdFile := filepath.Join(opt.FileDir, schemaLocation)
	var fi os.FileInfo
	fi, err = os.Stat(xsdFile)
	if err != nil {
		return
	}
	if fi.IsDir() {
		// extract type of value from include schema.
		valueType = ""
		for include := range opt.IncludeMap {
			parser := NewParser(&Options{
				FilePath:            filepath.Join(opt.FileDir, include),
				OutputDir:           opt.OutputDir,
				Extract:             true,
				Lang:                opt.Lang,
				IncludeMap:          opt.IncludeMap,
				LocalNameNSMap:      opt.LocalNameNSMap,
				NSSchemaLocationMap: opt.NSSchemaLocationMap,
				ParseFileList:       opt.ParseFileList,
				ParseFileMap:        opt.ParseFileMap,
				ProtoTree:           make([]interface{}, 0),
			})
			if parser.Parse() != nil {
				return
			}
			if vt := getBasefromSimpleType(trimNSPrefix(value), parser.ProtoTree); vt != trimNSPrefix(value) {
				valueType = vt
			}
		}
		if valueType == "" {
			valueType = trimNSPrefix(value)
		}
		return
	}

	depXSDSchema, ok := opt.ParseFileMap[xsdFile]
	if !ok {
		parser := NewParser(&Options{
			FilePath:            xsdFile,
			OutputDir:           opt.OutputDir,
			Extract:             false,
			Lang:                opt.Lang,
			IncludeMap:          opt.IncludeMap,
			LocalNameNSMap:      opt.LocalNameNSMap,
			NSSchemaLocationMap: opt.NSSchemaLocationMap,
			ParseFileList:       opt.ParseFileList,
			ParseFileMap:        opt.ParseFileMap,
			ProtoTree:           make([]interface{}, 0),
		})
		if parser.Parse() != nil {
			return
		}
		depXSDSchema = parser.ProtoTree
	}
	valueType = getBasefromSimpleType(trimNSPrefix(value), depXSDSchema)
	if valueType != trimNSPrefix(value) && valueType != "" {
		return
	}
	parser := NewParser(&Options{
		FilePath:            xsdFile,
		OutputDir:           opt.OutputDir,
		Extract:             true,
		Lang:                opt.Lang,
		IncludeMap:          opt.IncludeMap,
		LocalNameNSMap:      opt.LocalNameNSMap,
		NSSchemaLocationMap: opt.NSSchemaLocationMap,
		ParseFileList:       opt.ParseFileList,
		ParseFileMap:        opt.ParseFileMap,
		ProtoTree:           make([]interface{}, 0),
	})
	if parser.Parse() != nil {
		return
	}
	valueType = getBasefromSimpleType(trimNSPrefix(value), parser.ProtoTree)
	return
}
