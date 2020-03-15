// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"fmt"
	"os"
	"strings"
)

var RustBuildinType = map[string]bool{
	"i8":        true,
	"i16":       true,
	"i32":       true,
	"i64":       true,
	"i128":      true,
	"isize":     true,
	"u8":        true,
	"u16":       true,
	"u32":       true,
	"u64":       true,
	"u128":      true,
	"usize":     true,
	"f32":       true,
	"f64":       true,
	"Vec<char>": true,
	"Vec<u8>":   true,
	"&[u8]":     true,
	"bool":      true,
	"char":      true,
}

// GenRust generate Go programming language source code for XML schema
// definition files.
func (gen *CodeGenerator) GenRust() error {
	structAST := map[string]string{}
	var field string
	for _, ele := range gen.ProtoTree {
		switch v := ele.(type) {
		case *SimpleType:
			if v.List {
				if _, ok := structAST[v.Name]; !ok {
					fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
					content := fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", v.Name, genRustFieldName(v.Name), fieldType)
					structAST[v.Name] = content
					field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", genRustFieldName(v.Name), structAST[v.Name])
					continue
				}
			}
			if v.Union && len(v.MemberTypes) > 0 {
				if _, ok := structAST[v.Name]; !ok {
					var content string
					for memberName, memberType := range v.MemberTypes {
						if memberType == "" { // fix order issue
							memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
						}
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, genRustFieldName(memberName), genRustFieldType(memberType))
					}
					structAST[v.Name] = content
					field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", genRustFieldName(v.Name), structAST[v.Name])
				}
				continue
			}
			if _, ok := structAST[v.Name]; !ok {
				fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
				content := fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, genRustFieldName(v.Name), fieldType)
				structAST[v.Name] = content
				field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", genRustFieldName(v.Name), structAST[v.Name])
			}

		case *ComplexType:
			if _, ok := structAST[v.Name]; !ok {
				var content string
				for _, attrGroup := range v.AttributeGroup {
					fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
					content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", attrGroup.Name, genRustFieldName(attrGroup.Name), genRustFieldType(fieldType))
				}

				for _, attribute := range v.Attributes {
					// TODO: check attribute.Optional
					fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
					content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", attribute.Name, genRustFieldName(attribute.Name), fieldType)
				}
				for _, group := range v.Groups {
					fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree))
					fieldName := genRustFieldName(group.Name)
					if group.Plural {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", group.Name, fieldName, fieldType)
					} else {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", group.Name, fieldName, fieldType)
					}
				}
				for _, element := range v.Elements {
					fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
					fieldName := genRustFieldName(element.Name)
					if element.Plural {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", element.Name, fieldName, fieldType)
					} else {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", element.Name, fieldName, fieldType)
					}
				}
				structAST[v.Name] = content
				field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", genRustFieldName(v.Name), structAST[v.Name])
			}

		case *Group:
			if _, ok := structAST[v.Name]; !ok {
				var content string
				for _, element := range v.Elements {
					fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
					fieldName := genRustFieldName(element.Name)
					if v.Plural {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", element.Name, fieldName, fieldType)
					} else {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", element.Name, fieldName, fieldType)
					}
				}
				for _, group := range v.Groups {
					fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree))
					fieldName := genRustFieldName(group.Name)
					if v.Plural {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", group.Name, fieldName, fieldType)
					} else {
						content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", group.Name, fieldName, fieldType)
					}
				}
				structAST[v.Name] = content
				field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", genRustFieldName(v.Name), structAST[v.Name])
			}
		case *AttributeGroup:
			if _, ok := structAST[v.Name]; !ok {
				var content string
				for _, attribute := range v.Attributes {
					// TODO: check attribute.Optional
					content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", attribute.Name, genRustFieldName(attribute.Name), genRustFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)))
				}
				structAST[v.Name] = content
				field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", genRustFieldName(v.Name), structAST[v.Name])

			}
		case *Element:
			if _, ok := structAST[v.Name]; !ok {
				fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
				fieldName := genRustFieldName(v.Name)
				if v.Plural {
					structAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", v.Name, fieldName, fieldType)
				} else {
					structAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, fieldName, fieldType)
				}
				field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", fieldName, structAST[v.Name])
			}

		case *Attribute:
			if _, ok := structAST[v.Name]; !ok {
				fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
				fieldName := genRustFieldName(v.Name)
				if v.Plural {
					structAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", v.Name, fieldName, fieldType)
				} else {
					structAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, fieldName, fieldType)
				}
				field += fmt.Sprintf("\n#[derive(Debug, Serialize, Deserialize)]\nstruct %s {\n%s}\n", fieldName, structAST[v.Name])
			}
		}
	}
	f, err := os.Create(gen.File + ".rs")
	if err != nil {
		return err
	}
	defer f.Close()
	var extern = `#[macro_use]
extern crate serde_derive;
extern crate serde;
extern crate serde_xml_rs;

use serde_xml_rs::from_reader;`
	source := []byte(fmt.Sprintf("%s\n\n%s\n%s", copyright, extern, field))
	f.Write(source)
	return err
}

func genRustFieldName(name string) (fieldName string) {
	for _, str := range strings.Split(name, ":") {
		fieldName += MakeFirstUpperCase(str)
	}
	var tmp string
	for _, str := range strings.Split(fieldName, ".") {
		tmp += MakeFirstUpperCase(str)
	}
	fieldName = tmp
	fieldName = strings.Replace(fieldName, "-", "", -1)
	return
}

func genRustFieldType(name string) string {
	if _, ok := RustBuildinType[name]; ok {
		return name
	}
	var fieldType string
	for _, str := range strings.Split(name, ".") {
		fieldType += MakeFirstUpperCase(str)
	}
	fieldType = MakeFirstUpperCase(strings.Replace(fieldType, "-", "", -1))
	if fieldType != "" {
		return fieldType
	}
	return "char"
}
