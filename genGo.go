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
	"go/format"
	"os"
	"reflect"
	"strings"
)

// CodeGenerator holds code generator overrides and runtime data that are used
// when generate code from proto tree.
type CodeGenerator struct {
	Lang              string
	File              string
	Field             string
	Package           string
	ImportTime        bool // For Go language
	ImportEncodingXML bool // For Go language
	NoXMLName         bool // For Go language
	ProtoTree         []interface{}
	StructAST         map[string]string
}

var goBuildinType = map[string]bool{
	"xml.Name":      true,
	"byte":          true,
	"[]byte":        true,
	"bool":          true,
	"[]bool":        true,
	"complex64":     true,
	"complex128":    true,
	"float32":       true,
	"float64":       true,
	"int":           true,
	"int8":          true,
	"int16":         true,
	"int32":         true,
	"int64":         true,
	"interface":     true,
	"[]interface{}": true,
	"string":        true,
	"[]string":      true,
	"time.Time":     true,
	"uint":          true,
	"uint8":         true,
	"uint16":        true,
	"uint32":        true,
	"uint64":        true,
}

// GenGo generate Go programming language source code for XML schema
// definition files.
func (gen *CodeGenerator) GenGo() error {
	fieldNameCount = make(map[string]int)
	for _, ele := range gen.ProtoTree {
		if ele == nil {
			continue
		}
		funcName := fmt.Sprintf("Go%s", reflect.TypeOf(ele).String()[6:])
		callFuncByName(gen, funcName, []reflect.Value{reflect.ValueOf(ele)})
	}
	f, err := os.Create(gen.FileWithExtension(".go"))
	if err != nil {
		return err
	}
	defer f.Close()
	var importPackage, packages string
	if gen.ImportTime {
		packages += "\t\"time\"\n"
	}
	if gen.ImportEncodingXML {
		packages += "\t\"encoding/xml\"\n"
	}
	if packages != "" {
		importPackage = fmt.Sprintf("import (\n%s)", packages)
	}
	packageName := gen.Package
	if packageName == "" {
		packageName = "schema"
	}
	source, err := format.Source([]byte(fmt.Sprintf("%s\n\npackage %s\n%s%s", copyright, packageName, importPackage, gen.Field)))
	if err != nil {
		f.WriteString(fmt.Sprintf("package %s\n%s%s", packageName, importPackage, gen.Field))
		return err
	}
	f.Write(source)
	return err
}

func genGoFieldName(name string, unique bool) (fieldName string) {
	for _, str := range strings.Split(name, ":") {
		fieldName += MakeFirstUpperCase(str)
	}
	var tmp string
	for _, str := range strings.Split(fieldName, ".") {
		tmp += MakeFirstUpperCase(str)
	}
	fieldName = tmp
	fieldName = strings.Replace(strings.Replace(fieldName, "-", "", -1), "_", "", -1)
	if unique {
		fieldNameCount[fieldName]++
		if count := fieldNameCount[fieldName]; count != 1 {
			fieldName = fmt.Sprintf("%s%d", fieldName, count)
		}
	}
	return
}

func genGoFieldType(name string) string {
	if _, ok := goBuildinType[name]; ok {
		return name
	}
	var fieldType string
	for _, str := range strings.Split(name, ".") {
		fieldType += MakeFirstUpperCase(str)
	}
	fieldType = strings.Replace(MakeFirstUpperCase(strings.Replace(fieldType, "-", "", -1)), "_", "", -1)
	if fieldType != "" {
		return "*" + fieldType
	}
	return "interface{}"
}

// GoSimpleType generates code for simple type XML schema in Go language
// syntax.
func (gen *CodeGenerator) GoSimpleType(v *SimpleType) {
	if v.List {
		if _, ok := gen.StructAST[v.Name]; !ok {
			fieldType := genGoFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			if fieldType == "time.Time" {
				gen.ImportTime = true
			}
			content := fmt.Sprintf(" []%s\n", genGoFieldType(fieldType))
			gen.StructAST[v.Name] = content
			fieldName := genGoFieldName(v.Name, true)
			gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
			return
		}
	}
	if v.Union && len(v.MemberTypes) > 0 {
		if _, ok := gen.StructAST[v.Name]; !ok {
			content := " struct {\n"
			fieldName := genGoFieldName(v.Name, true)
			if fieldName != v.Name && !gen.NoXMLName {
				gen.ImportEncodingXML = true
				content += fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`\n", v.Name)
			}
			for _, member := range toSortedPairs(v.MemberTypes) {
				memberName := member.key
				memberType := member.value

				if memberType == "" { // fix order issue
					memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
				}
				content += fmt.Sprintf("\t%s\t%s\n", genGoFieldName(memberName, false), genGoFieldType(memberType))
			}
			content += "}\n"
			gen.StructAST[v.Name] = content
			gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
		}
		return
	}
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := fmt.Sprintf(" %s\n", genGoFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree)))
		gen.StructAST[v.Name] = content
		fieldName := genGoFieldName(v.Name, true)
		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

// GoComplexType generates code for complex type XML schema in Go language
// syntax.
func (gen *CodeGenerator) GoComplexType(v *ComplexType) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " struct {\n"
		fieldName := genGoFieldName(v.Name, true)
		if fieldName != v.Name && !gen.NoXMLName {
			gen.ImportEncodingXML = true
			content += fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`\n", v.Name)
		}
		for _, attrGroup := range v.AttributeGroup {
			fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
			if fieldType == "time.Time" {
				gen.ImportTime = true
			}
			content += fmt.Sprintf("\t%s\t%s\n", genGoFieldName(attrGroup.Name, false), genGoFieldType(fieldType))
		}

		for _, attribute := range v.Attributes {
			var optional string
			if attribute.Optional {
				optional = `,omitempty`
			}
			fieldType := genGoFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
			if fieldType == "time.Time" {
				gen.ImportTime = true
			}
			content += fmt.Sprintf("\t%sAttr\t%s\t`xml:\"%s,attr%s\"`\n", genGoFieldName(attribute.Name, false), fieldType, attribute.Name, optional)
		}
		for _, group := range v.Groups {
			var plural string
			if group.Plural {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s\t%s%s\n", genGoFieldName(group.Name, false), plural, genGoFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
		}

		for _, element := range v.Elements {
			var plural string
			if element.Plural {
				plural = "[]"
			}
			fieldType := genGoFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
			if fieldType == "time.Time" {
				gen.ImportTime = true
			}
			content += fmt.Sprintf("\t%s\t%s%s\t`xml:\"%s\"`\n", genGoFieldName(element.Name, false), plural, fieldType, element.Name)
		}
		if len(v.Base) > 0 {
			// If the type is a built-in type, generate a Value field as chardata.
			// If it's not built-in one, embed the base type in the struct for the child type
			// to effectively inherit all of the base type's fields
			if isGoBuiltInType(v.Base) {
				content += fmt.Sprintf("\tValue\t%s\t`xml:\",chardata\"`\n", genGoFieldType(v.Base))
			} else {
				content += fmt.Sprintf("\t%s\n", genGoFieldType(v.Base))
			}
		}
		content += "}\n"
		gen.StructAST[v.Name] = content
		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

func isGoBuiltInType(typeName string) bool {
	_, builtIn := goBuildinType[typeName]
	return builtIn
}

// GoGroup generates code for group XML schema in Go language syntax.
func (gen *CodeGenerator) GoGroup(v *Group) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " struct {\n"
		fieldName := genGoFieldName(v.Name, true)
		if fieldName != v.Name && !gen.NoXMLName{
			gen.ImportEncodingXML = true
			content += fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`\n", v.Name)
		}
		for _, element := range v.Elements {
			var plural string
			if element.Plural {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s\t%s%s\n", genGoFieldName(element.Name, false), plural, genGoFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)))
		}

		for _, group := range v.Groups {
			var plural string
			if group.Plural {
				plural = "[]"
			}
			content += fmt.Sprintf("\t%s\t%s%s\n", genGoFieldName(group.Name, false), plural, genGoFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
		}

		content += "}\n"
		gen.StructAST[v.Name] = content
		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

// GoAttributeGroup generates code for attribute group XML schema in Go language
// syntax.
func (gen *CodeGenerator) GoAttributeGroup(v *AttributeGroup) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " struct {\n"
		fieldName := genGoFieldName(v.Name, true)
		if fieldName != v.Name && !gen.NoXMLName{
			gen.ImportEncodingXML = true
			content += fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`\n", v.Name)
		}
		for _, attribute := range v.Attributes {
			var optional string
			if attribute.Optional {
				optional = `,omitempty`
			}
			content += fmt.Sprintf("\t%sAttr\t%s\t`xml:\"%s,attr%s\"`\n", genGoFieldName(attribute.Name, false), genGoFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)), attribute.Name, optional)
		}
		content += "}\n"
		gen.StructAST[v.Name] = content
		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

// GoElement generates code for element XML schema in Go language syntax.
func (gen *CodeGenerator) GoElement(v *Element) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		var plural string
		if v.Plural {
			plural = "[]"
		}
		content := fmt.Sprintf("\t%s%s\n", plural, genGoFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
		gen.StructAST[v.Name] = content
		fieldName := genGoFieldName(v.Name, false)
		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

// GoAttribute generates code for attribute XML schema in Go language syntax.
func (gen *CodeGenerator) GoAttribute(v *Attribute) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		var plural string
		if v.Plural {
			plural = "[]"
		}
		content := fmt.Sprintf("\t%s%s\n", plural, genGoFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
		gen.StructAST[v.Name] = content
		fieldName := genGoFieldName(v.Name, true)
		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

func (gen *CodeGenerator) FileWithExtension(extension string) string {
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	if strings.HasSuffix(gen.File, extension) {
		return gen.File
	}
	return gen.File + extension
}
