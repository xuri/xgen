// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

var (
	rustBuildinType = map[string]bool{
		"i8":          true,
		"i16":         true,
		"i32":         true,
		"i64":         true,
		"i128":        true,
		"isize":       true,
		"u8":          true,
		"u16":         true,
		"u32":         true,
		"u64":         true,
		"u128":        true,
		"usize":       true,
		"f32":         true,
		"f64":         true,
		"Vec<char>":   true,
		"Vec<String>": true,
		"Vec<u8>":     true,
		"bool":        true,
		"char":        true,
		"String":      true,
	}
	rustKeywords = map[string]bool{
		"as":       true,
		"break":    true,
		"const":    true,
		"continue": true,
		"crate":    true,
		"dyn":      true,
		"else":     true,
		"enum":     true,
		"extern":   true,
		"false":    true,
		"fn":       true,
		"for":      true,
		"if":       true,
		"impl":     true,
		"in":       true,
		"let":      true,
		"loop":     true,
		"match":    true,
		"mod":      true,
		"move":     true,
		"mut":      true,
		"pub":      true,
		"ref":      true,
		"return":   true,
		"Self":     true,
		"self":     true,
		"static":   true,
		"struct":   true,
		"super":    true,
		"trait":    true,
		"true":     true,
		"type":     true,
		"unsafe":   true,
		"use":      true,
		"where":    true,
		"while":    true,
		"abstract": true,
		"async":    true,
		"await":    true,
		"become":   true,
		"box":      true,
		"do":       true,
		"final":    true,
		"macro":    true,
		"override": true,
		"priv":     true,
		"try":      true,
		"typeof":   true,
		"unsized":  true,
		"virtual":  true,
		"yield":    true,
	}
)

// GenRust generate Go programming language source code for XML schema
// definition files.
func (gen *CodeGenerator) GenRust() error {
	for _, ele := range gen.ProtoTree {
		if ele == nil {
			continue
		}
		funcName := fmt.Sprintf("Rust%s", reflect.TypeOf(ele).String()[6:])
		callFuncByName(gen, funcName, []reflect.Value{reflect.ValueOf(ele)})
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
	source := []byte(fmt.Sprintf("%s\n\n%s\n%s", copyright, extern, gen.Field))
	f.Write(source)
	return err
}

// genRustFieldName generate struct field name for Rust code.
func genRustFieldName(name string) (fieldName string) {
	for _, str := range strings.Split(name, ":") {
		fieldName += MakeFirstUpperCase(str)
	}
	var tmp string
	for _, str := range strings.Split(fieldName, ".") {
		tmp += MakeFirstUpperCase(str)
	}
	fieldName = tmp
	fieldName = ToSnakeCase(strings.Replace(fieldName, "-", "", -1))
	if _, ok := rustKeywords[fieldName]; ok {
		fieldName += "_attr"
	}
	return
}

// genRustStructName generate struct name for Rust code.
func genRustStructName(name string) (structName string) {
	for _, str := range strings.Split(name, ":") {
		structName += MakeFirstUpperCase(str)
	}
	var tmp string
	for _, str := range strings.Split(structName, ".") {
		tmp += MakeFirstUpperCase(str)
	}
	structName = tmp
	structName = strings.NewReplacer("-", "", "_", "").Replace(structName)
	return
}

// genRustFieldType generate struct field type for Rust code.
func genRustFieldType(name string) string {
	if _, ok := rustBuildinType[name]; ok {
		return name
	}
	fieldType := genRustStructName(name)
	if fieldType != "" {
		return fieldType
	}
	return "char"
}

// RustSimpleType generates code for simple type XML schema in Rust language
// syntax.
func (gen *CodeGenerator) RustSimpleType(v *SimpleType) {
	if v.List {
		if _, ok := gen.StructAST[v.Name]; !ok {
			fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			content := fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", v.Name, genRustFieldName(v.Name), fieldType)
			gen.StructAST[v.Name] = content
			fieldName := genRustStructName(v.Name)
			gen.Field += fmt.Sprintf("\n%s#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
			return
		}
	}
	if v.Union && len(v.MemberTypes) > 0 {
		if _, ok := gen.StructAST[v.Name]; !ok {
			var content string
			for memberName, memberType := range v.MemberTypes {
				if memberType == "" { // fix order issue
					memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
				}
				content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, genRustFieldName(memberName), genRustFieldType(memberType))
			}
			gen.StructAST[v.Name] = content
			gen.Field += fmt.Sprintf("\n#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genRustStructName(v.Name), gen.StructAST[v.Name])
		}
		return
	}
	if _, ok := gen.StructAST[v.Name]; !ok {
		fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
		content := fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, genRustFieldName(v.Name), fieldType)
		gen.StructAST[v.Name] = content
		fieldName := genRustStructName(v.Name)
		gen.Field += fmt.Sprintf("\n%s#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// RustComplexType generates code for complex type XML schema in Rust language
// syntax.
func (gen *CodeGenerator) RustComplexType(v *ComplexType) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		var content string
		for _, attrGroup := range v.AttributeGroup {
			fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
			content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", attrGroup.Name, genRustFieldName(attrGroup.Name), genRustFieldType(fieldType))
		}
		for _, attribute := range v.Attributes {
			fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
			if attribute.Optional {
				content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Option<%s>,\n", attribute.Name, genRustFieldName(attribute.Name), fieldType)
			} else {
				content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", attribute.Name, genRustFieldName(attribute.Name), fieldType)
			}
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
				if element.Optional {
					content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Option<%s>,\n", element.Name, fieldName, fieldType)
				} else {
					content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", element.Name, fieldName, fieldType)
				}
			}

		}
		gen.StructAST[v.Name] = content
		fieldName := genRustStructName(v.Name)
		gen.Field += fmt.Sprintf("\n%s#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// RustGroup generates code for group XML schema in Rust language syntax.
func (gen *CodeGenerator) RustGroup(v *Group) {
	if _, ok := gen.StructAST[v.Name]; !ok {
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
		gen.StructAST[v.Name] = content
		fieldName := genRustStructName(v.Name)
		gen.Field += fmt.Sprintf("\n%s#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// RustAttributeGroup generates code for attribute group XML schema in Rust language
// syntax.
func (gen *CodeGenerator) RustAttributeGroup(v *AttributeGroup) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		var content string
		for _, attribute := range v.Attributes {
			if attribute.Optional {
				content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Option<%s>,\n", attribute.Name, genRustFieldName(attribute.Name), genRustFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)))
			} else {
				content += fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", attribute.Name, genRustFieldName(attribute.Name), genRustFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)))
			}
		}
		gen.StructAST[v.Name] = content
		fieldName := genRustStructName(v.Name)
		gen.Field += fmt.Sprintf("\n%s#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// RustElement generates code for element XML schema in Rust language syntax.
func (gen *CodeGenerator) RustElement(v *Element) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
		fieldName := genRustFieldName(v.Name)
		if v.Plural {
			gen.StructAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", v.Name, fieldName, fieldType)
		} else {
			gen.StructAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, fieldName, fieldType)
		}
		gen.Field += fmt.Sprintf("\n%s#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// RustAttribute generates code for attribute XML schema in Rust language syntax.
func (gen *CodeGenerator) RustAttribute(v *Attribute) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		fieldType := genRustFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
		fieldName := genRustFieldName(v.Name)
		if v.Plural {
			gen.StructAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: Vec<%s>,\n", v.Name, fieldName, fieldType)
		} else {
			gen.StructAST[v.Name] = fmt.Sprintf("\t#[serde(rename = \"%s\")]\n\tpub %s: %s,\n", v.Name, fieldName, fieldType)
		}
		gen.Field += fmt.Sprintf("\n%s#[derive(Debug, Deserialize, Serialize, PartialEq)]\npub struct %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}
