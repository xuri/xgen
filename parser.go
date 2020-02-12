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
func (opt *Options) Parse() (protoTree []interface{}, err error) {
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
		opt.ParseFileMap[opt.FilePath] = protoTree
	}
	protoTree = make([]interface{}, 0)

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
				rt := onEleFunc.Call([]reflect.Value{reflect.ValueOf(element), reflect.ValueOf(protoTree)})
				if !rt[0].IsNil() {
					err = rt[0].Interface().(error)
					return
				}
			}

		case xml.EndElement:
			if opt.ComplexType.Len() > 0 {
				if element.Name.Local == opt.CurrentEle && opt.ComplexType.Len() == 1 {
					protoTree = append(protoTree, opt.ComplexType.Pop())
					opt.CurrentEle = ""
					continue
				}
				if element.Name.Local == "complexType" {
					protoTree = append(protoTree, opt.ComplexType.Pop())
					opt.CurrentEle = ""
					continue
				}
			}
			if opt.SimpleType.Len() > 0 {
				if element.Name.Local == opt.CurrentEle && !opt.InUnion {
					protoTree = append(protoTree, opt.SimpleType.Pop())
					opt.CurrentEle = ""
				}
				if element.Name.Local == "union" {
					opt.InUnion = false
				}
				if opt.Element.Len() > 0 {
					opt.Element.Peek().(*Element).Type, err = opt.GetValueType(opt.SimpleType.Pop().(*SimpleType).Base, protoTree)
					if err != nil {
						return
					}
					opt.CurrentEle = ""
				}
				if opt.Attribute != nil && opt.SimpleType.Peek() != nil {
					opt.Attribute.Type, err = opt.GetValueType(opt.SimpleType.Pop().(*SimpleType).Base, protoTree)
					if err != nil {
						return
					}
					opt.CurrentEle = ""
				}
			}
			if opt.Attribute != nil && opt.ComplexType.Len() == 0 {
				protoTree = append(protoTree, opt.Attribute)
			}
			if element.Name.Local == "element" {
				if opt.Element.Len() > 0 && opt.ComplexType.Len() == 0 {
					protoTree = append(protoTree, opt.Element.Pop())
				}
			}

			if opt.Group != nil {
				if element.Name.Local == opt.CurrentEle && opt.InGroup == 1 {
					protoTree = append(protoTree, opt.Group)
					opt.CurrentEle = ""
					opt.InGroup--
					opt.Group = nil
				}
				if element.Name.Local == opt.CurrentEle {
					opt.InGroup--
				}

			}
			if opt.AttributeGroup != nil {
				if element.Name.Local == opt.CurrentEle && opt.InAttributeGroup {
					protoTree = append(protoTree, opt.AttributeGroup)
					opt.CurrentEle = ""
					opt.InAttributeGroup = false
					opt.AttributeGroup = nil
				}
			}
		default:
		}
	}
	defer xmlFile.Close()

	if !opt.Extract {
		opt.ParseFileList[opt.FilePath] = true
		opt.ParseFileMap[opt.FilePath] = protoTree
		if err = genCode(filepath.Join(opt.OutputDir, filepath.Base(opt.FilePath)), protoTree); err != nil {
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
		depXSDSchema, err = NewParser(&Options{
			FilePath:            xsdFile,
			OutputDir:           opt.OutputDir,
			Extract:             false,
			LocalNameNSMap:      opt.LocalNameNSMap,
			NSSchemaLocationMap: opt.NSSchemaLocationMap,
			ParseFileList:       opt.ParseFileList,
			ParseFileMap:        opt.ParseFileMap,
		}).Parse()
		if err != nil {
			return
		}
	}
	valueType = getBasefromSimpleType(trimNSPrefix(value), depXSDSchema)
	if valueType != trimNSPrefix(value) && valueType != "" {
		return
	}
	extractXSDSchema, err := NewParser(&Options{
		FilePath:            xsdFile,
		OutputDir:           opt.OutputDir,
		Extract:             true,
		LocalNameNSMap:      opt.LocalNameNSMap,
		NSSchemaLocationMap: opt.NSSchemaLocationMap,
		ParseFileList:       opt.ParseFileList,
		ParseFileMap:        opt.ParseFileMap,
	}).Parse()
	if err != nil {
		return
	}
	valueType = getBasefromSimpleType(trimNSPrefix(value), extractXSDSchema)
	return
}
