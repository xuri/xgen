// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
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

	"golang.org/x/net/html/charset"
)

// Options holds user-defined overrides and runtime data that are used when
// parsing from an XSD document.
type Options struct {
	FilePath            string
	FileDir             string
	OutputDir           string
	Extract             bool
	LocalNameNSMap      map[string]string
	NSSchemaLocationMap map[string]string
	ParseFileList       map[string]bool
	ParseFileMap        map[string][]interface{}
	ProtoTree           []interface{}

	InElement        string
	CurrentEle       string
	InGroup          int
	InUnion          bool
	InAttributeGroup bool

	SimpleType     *Stack
	ComplexType    *Stack
	Element        *Stack
	Attribute      *Attribute
	Group          *Group
	AttributeGroup *AttributeGroup
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
	opt.Attribute = nil
	opt.Group = nil
	opt.AttributeGroup = nil

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
			onEleFunc := reflect.ValueOf(opt).MethodByName(funcName)
			if onEleFunc.IsValid() {
				rt := onEleFunc.Call([]reflect.Value{reflect.ValueOf(element), reflect.ValueOf(opt.ProtoTree)})
				if !rt[0].IsNil() {
					err = rt[0].Interface().(error)
					return
				}
			}

		case xml.EndElement:
			funcName := fmt.Sprintf("End%s", MakeFirstUpperCase(element.Name.Local))
			onEleFunc := reflect.ValueOf(opt).MethodByName(funcName)
			if onEleFunc.IsValid() {
				rt := onEleFunc.Call([]reflect.Value{reflect.ValueOf(element), reflect.ValueOf(opt.ProtoTree)})
				if !rt[0].IsNil() {
					err = rt[0].Interface().(error)
					return
				}
			}
		default:
		}

	}
	defer xmlFile.Close()

	if !opt.Extract {
		opt.ParseFileList[opt.FilePath] = true
		opt.ParseFileMap[opt.FilePath] = opt.ProtoTree
		if err = genCode(filepath.Join(opt.OutputDir, filepath.Base(opt.FilePath)), opt.ProtoTree); err != nil {
			return
		}
	}
	return
}

// GetValueType convert XSD schema value type to the build-in type for the
// given value and proto tree.
func (opt *Options) GetValueType(value string, XSDSchema []interface{}) (valueType string, err error) {
	if buildType, ok := buildInTypes[trimNSPrefix(value)]; ok {
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
	xsdFile := filepath.Join(opt.FileDir, schemaLocation)
	var fi os.FileInfo
	fi, err = os.Stat(xsdFile)
	if err != nil {
		return
	}
	if fi.IsDir() {
		return
	}

	depXSDSchema, ok := opt.ParseFileMap[xsdFile]
	if !ok {
		parser := NewParser(&Options{
			FilePath:            xsdFile,
			OutputDir:           opt.OutputDir,
			Extract:             false,
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
