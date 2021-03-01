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

var cBuildInType = map[string]bool{
	"bool":           true,
	"char":           true,
	"unsigned char":  true,
	"signed char":    true,
	"char[]":         true, // char[] will be flat to 'char field_name[]'
	"float":          true,
	"double":         true,
	"long double":    true,
	"int":            true,
	"unsigned int":   true,
	"short":          true,
	"unsigned short": true,
	"long":           true,
	"unsigned long":  true,
	"void":           true,
	"enum":           true,
}

// GenC generates C programming language source code for XML schema definition
// files.
func (gen *CodeGenerator) GenC() error {
	for _, ele := range gen.ProtoTree {
		if ele == nil {
			continue
		}
		funcName := fmt.Sprintf("C%s", reflect.TypeOf(ele).String()[6:])
		callFuncByName(gen, funcName, []reflect.Value{reflect.ValueOf(ele)})
	}
	f, err := os.Create(gen.File + ".h")
	if err != nil {
		return err
	}
	defer f.Close()
	source := []byte(fmt.Sprintf("%s\n%s", copyright, gen.Field))
	f.Write(source)
	return err
}

func innerArray(dataType string) (string, bool) {
	if strings.HasSuffix(dataType, "[]") {
		return strings.TrimSuffix(dataType, "[]"), true
	}
	return dataType, false
}

func genCFieldName(name string) (fieldName string) {
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

func genCFieldType(name string) string {
	if _, ok := cBuildInType[name]; ok {
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
	return "void"
}

// CSimpleType generates code for simple type XML schema in C language
// syntax.
func (gen *CodeGenerator) CSimpleType(v *SimpleType) {
	if v.List {
		if _, ok := gen.StructAST[v.Name]; !ok {
			fieldType := genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			content := fmt.Sprintf("%s %s[];\n", genCFieldType(fieldType), genCFieldName(v.Name))
			gen.StructAST[v.Name] = content
			fieldName := genCFieldName(v.Name)
			gen.Field += fmt.Sprintf("%stypedef %s", genFieldComment(fieldName, v.Doc, "//"), gen.StructAST[v.Name])
			return
		}
	}
	if v.Union && len(v.MemberTypes) > 0 {
		if _, ok := gen.StructAST[v.Name]; !ok {
			content := "struct {\n"
			for memberName, memberType := range v.MemberTypes {
				if memberType == "" { // fix order issue
					memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
				}
				var plural, fieldType string
				var ok bool
				if fieldType, ok = innerArray(genCFieldType(memberType)); ok {
					plural = "[]"
				}
				content += fmt.Sprintf("\t%s %s%s;\n", fieldType, genCFieldName(memberName), plural)
			}
			content += "}"
			gen.StructAST[v.Name] = content
			fieldName := genCFieldName(v.Name)
			gen.Field += fmt.Sprintf("%stypedef %s %s;\n", genFieldComment(fieldName, v.Doc, "//"), gen.StructAST[v.Name], fieldName)
		}
		return
	}
	if _, ok := gen.StructAST[v.Name]; !ok {
		var plural, fieldType string
		var ok bool
		if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))); ok {
			plural = "[]"
		}
		gen.StructAST[v.Name] = fmt.Sprintf("%s %s%s", fieldType, genCFieldName(v.Name), plural)
		fieldName := genCFieldName(v.Name)
		gen.Field += fmt.Sprintf("%stypedef %s;\n", genFieldComment(fieldName, v.Doc, "//"), gen.StructAST[v.Name])
	}
	return
}

// CComplexType generates code for complex type XML schema in C language
// syntax.
func (gen *CodeGenerator) CComplexType(v *ComplexType) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := "struct {\n"
		for _, attrGroup := range v.AttributeGroup {
			fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
			content += fmt.Sprintf("\t%s %s;\n", genCFieldType(fieldType), genCFieldName(attrGroup.Name))
		}

		for _, attribute := range v.Attributes {
			var optional string
			if attribute.Optional {
				optional = `, optional`
			}
			var plural, fieldType string
			var ok bool
			if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))); ok {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s %sAttr%s; // attr%s\n", fieldType, genCFieldName(attribute.Name), plural, optional)
		}

		for _, group := range v.Groups {
			var plural string
			if group.Plural {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s %s%s;\n", genCFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)), genCFieldName(group.Name), plural)
		}

		for _, element := range v.Elements {
			var plural, fieldType string
			var ok bool
			if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))); ok || element.Plural {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s %s%s;\n", fieldType, genCFieldName(element.Name), plural)
		}

		if len(v.Base) > 0 {
			var plural, fieldType string
			var ok bool
			if fieldType, ok = innerArray(genCFieldType(v.Base)); ok {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s Value%s;\n", fieldType, plural)
		}
		content += "}"
		gen.StructAST[v.Name] = content
		fieldName := genCFieldName(v.Name)
		gen.Field += fmt.Sprintf("%stypedef %s %s;\n", genFieldComment(fieldName, v.Doc, "//"), gen.StructAST[v.Name], fieldName)
	}
	return
}

// CGroup generates code for group XML schema in C language syntax.
func (gen *CodeGenerator) CGroup(v *Group) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := "struct {\n"
		for _, element := range v.Elements {
			var plural string
			if element.Plural {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s %s%s;\n", genCFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)), genCFieldName(element.Name), plural)
		}

		for _, group := range v.Groups {
			var plural string
			if group.Plural {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s %s%s;\n", genCFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)), genCFieldName(group.Name), plural)
		}

		content += "}"
		gen.StructAST[v.Name] = content
		fieldName := genCFieldName(v.Name)
		gen.Field += fmt.Sprintf("%stypedef %s %s;\n", genFieldComment(fieldName, v.Doc, "//"), gen.StructAST[v.Name], fieldName)
	}
	return
}

// CAttributeGroup generates code for attribute group XML schema in C language
// syntax.
func (gen *CodeGenerator) CAttributeGroup(v *AttributeGroup) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := "struct {\n"
		for _, attribute := range v.Attributes {
			var optional, plural, fieldType string
			var ok bool
			if attribute.Optional {
				optional = `, optional`
			}
			if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))); ok {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s %sAttr%s; // attr%s\n", fieldType, genCFieldName(attribute.Name), plural, optional)
		}
		content += "}"
		gen.StructAST[v.Name] = content
		fieldName := genCFieldName(v.Name)
		gen.Field += fmt.Sprintf("%stypedef %s %s;\n", genFieldComment(fieldName, v.Doc, "//"), gen.StructAST[v.Name], fieldName)
	}
}

// CElement generates code for element XML schema in C language syntax.
func (gen *CodeGenerator) CElement(v *Element) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		var plural, fieldType string
		var ok bool
		if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))); ok || v.Plural {
			plural = "[]"
		}
		gen.StructAST[v.Name] = fmt.Sprintf("%s %s%s", fieldType, genCFieldName(v.Name), plural)
		gen.Field += fmt.Sprintf("\ntypedef %s;\n", gen.StructAST[v.Name])
	}
}

// CAttribute generates code for attribute XML schema in C language syntax.
func (gen *CodeGenerator) CAttribute(v *Attribute) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		var plural, fieldType string
		var ok bool
		if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))); ok || v.Plural {
			plural = "[]"
		}
		gen.StructAST[v.Name] = fmt.Sprintf("%s %s%s", fieldType, genCFieldName(v.Name), plural)
		fieldName := genCFieldName(v.Name)
		gen.Field += fmt.Sprintf("%stypedef %s;\n", genFieldComment(fieldName, v.Doc, "//"), gen.StructAST[v.Name])
	}
}
