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

var typeScriptBuildInType = map[string]bool{
	"boolean":    true,
	"number":     true,
	"string":     true,
	"void":       true,
	"null":       true,
	"undefined":  true,
	"Uint8Array": true,
}

// GenTypeScript generate TypeScript programming language source code for XML
// schema definition files.
func (gen *CodeGenerator) GenTypeScript() error {
	for _, ele := range gen.ProtoTree {
		if ele == nil {
			continue
		}
		funcName := fmt.Sprintf("TypeScript%s", reflect.TypeOf(ele).String()[6:])
		callFuncByName(gen, funcName, []reflect.Value{reflect.ValueOf(ele)})
	}
	f, err := os.Create(gen.File + ".ts")
	if err != nil {
		return err
	}
	defer f.Close()
	source := []byte(fmt.Sprintf("%s\n%s", copyright, gen.Field))
	f.Write(source)
	return err

}

func genTypeScriptFieldName(name string) (fieldName string) {
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

func genTypeScriptFieldType(name string, plural bool) (fieldType string) {
	if _, ok := typeScriptBuildInType[name]; ok {
		fieldType = name
		return
	}
	for _, str := range strings.Split(name, ".") {
		fieldType += MakeFirstUpperCase(str)
	}
	fieldType = MakeFirstUpperCase(strings.Replace(fieldType, "-", "", -1))
	if fieldType == "" || fieldType == "Any" {
		fieldType = "any"
	}
	if plural {
		fieldType = fmt.Sprintf("Array<%s>", fieldType)
	}
	return
}

// TypeScriptSimpleType generates code for simple type XML schema in TypeScript language
// syntax.
func (gen *CodeGenerator) TypeScriptSimpleType(v *SimpleType) {
	if v.List {
		if _, ok := gen.StructAST[v.Name]; !ok {
			fieldType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree), true)
			content := fmt.Sprintf(" = %s;\n", fieldType)
			gen.StructAST[v.Name] = content
			fieldName := genTypeScriptFieldName(v.Name)
			gen.Field += fmt.Sprintf("%sexport type %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
			return
		}
	}
	if v.Union && len(v.MemberTypes) > 0 {
		if _, ok := gen.StructAST[v.Name]; !ok {
			content := " {\n"
			for memberName, memberType := range v.MemberTypes {
				if memberType == "" { // fix order issue
					memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
				}
				content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(memberName), genTypeScriptFieldType(memberType, false))
			}
			content += "}\n"
			gen.StructAST[v.Name] = content
			fieldName := genTypeScriptFieldName(v.Name)
			gen.Field += fmt.Sprintf("%sexport class %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
		}
		return
	}
	if len(v.Restriction.Enum) > 0 {
		var content string
		baseType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree), false)
		for _, enum := range v.Restriction.Enum {
			switch baseType {
			case "string":
				content += fmt.Sprintf("\t%s = '%s',\n", enum, enum)
			case "number":
				content += fmt.Sprintf("\tEnum%s = %s,\n", enum, enum)
			default:
				content += fmt.Sprintf("\tEnum%s = '%s',\n", enum, enum)
			}
		}
		fieldName := genTypeScriptFieldName(v.Name)
		gen.Field += fmt.Sprintf("%sexport enum %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), fieldName, content)
		return
	}
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := fmt.Sprintf(" %s;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree), false))
		gen.StructAST[v.Name] = content
		fieldName := genTypeScriptFieldName(v.Name)
		gen.Field += fmt.Sprintf("%sexport type %s =%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// TypeScriptComplexType generates code for complex type XML schema in TypeScript language
// syntax.
func (gen *CodeGenerator) TypeScriptComplexType(v *ComplexType) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " {\n"
		for _, attrGroup := range v.AttributeGroup {
			fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
			content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(attrGroup.Name), genTypeScriptFieldType(fieldType, false))
		}

		for _, attribute := range v.Attributes {
			var optional string
			if attribute.Optional {
				optional = ` | null`
			}
			fieldType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree), attribute.Plural)
			content += fmt.Sprintf("\t%sAttr: %s%s;\n", genTypeScriptFieldName(attribute.Name), fieldType, optional)
		}
		for _, group := range v.Groups {
			content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(group.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree), group.Plural))
		}

		for _, element := range v.Elements {
			fieldType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree), element.Plural)
			content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(element.Name), fieldType)
		}

		if len(v.Base) > 0 {
			fieldType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			content += fmt.Sprintf("\tValue: %s;\n", fieldType)
		}
		content += "}\n"
		gen.StructAST[v.Name] = content
		fieldName := genTypeScriptFieldName(v.Name)
		gen.Field += fmt.Sprintf("%sexport class %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// TypeScriptGroup generates code for group XML schema in TypeScript language syntax.
func (gen *CodeGenerator) TypeScriptGroup(v *Group) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " {\n"
		for _, element := range v.Elements {
			content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(element.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree), element.Plural))
		}

		for _, group := range v.Groups {
			content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(group.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree), group.Plural))
		}

		content += "}\n"
		gen.StructAST[v.Name] = content
		fieldName := genTypeScriptFieldName(v.Name)
		gen.Field += fmt.Sprintf("%sexport class %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// TypeScriptAttributeGroup generates code for attribute group XML schema in TypeScript language
// syntax.
func (gen *CodeGenerator) TypeScriptAttributeGroup(v *AttributeGroup) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " {\n"
		for _, attribute := range v.Attributes {
			var optional string
			if attribute.Optional {
				optional = ` | null`
			}
			content += fmt.Sprintf("\t%sAttr: %s%s;\n", genTypeScriptFieldName(attribute.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree), attribute.Plural), optional)
		}
		content += "}\n"
		gen.StructAST[v.Name] = content
		fieldName := genTypeScriptFieldName(v.Name)
		gen.Field += fmt.Sprintf("%sexport class %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// TypeScriptElement generates code for element XML schema in TypeScript language syntax.
func (gen *CodeGenerator) TypeScriptElement(v *Element) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		gen.StructAST[v.Name] = fmt.Sprintf(" %s;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree), v.Plural))
		fieldName := genTypeScriptFieldName(v.Name)
		gen.Field += fmt.Sprintf("%sexport type %s =%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}

// TypeScriptAttribute generates code for attribute XML schema in TypeScript language syntax.
func (gen *CodeGenerator) TypeScriptAttribute(v *Attribute) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		gen.StructAST[v.Name] = fmt.Sprintf(" %s;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree), v.Plural))
		fieldName := genTypeScriptFieldName(v.Name)
		gen.Field += fmt.Sprintf("%sexport type %s =%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
	return
}
