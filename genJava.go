// Copyright 2020 - 2024 The xgen Authors. All rights reserved. Use of this
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

var javaBuildInType = map[string]bool{
	"Boolean":      true,
	"Byte":         true,
	"Character":    true,
	"List<String>": true,
	"List<Byte>":   true,
	"Float":        true,
	"Integer":      true,
	"Short":        true,
	"String":       true,
	"QName":        true,
	"Long":         true,
}

// GenJava generate Java programming language source code for XML schema
// definition files.
func (gen *CodeGenerator) GenJava() error {
	fieldNameCount = make(map[string]int)
	for _, ele := range gen.ProtoTree {
		if ele == nil {
			continue
		}
		funcName := fmt.Sprintf("Java%s", reflect.TypeOf(ele).String()[6:])
		callFuncByName(gen, funcName, []reflect.Value{reflect.ValueOf(ele)})
	}
	f, err := os.Create(gen.FileWithExtension(".java"))
	if err != nil {
		return err
	}
	defer f.Close()
	packageName := gen.Package
	if packageName == "" {
		packageName = "schema"
	}
	importPackage := `import java.util.ArrayList;
import java.util.List;
import javax.xml.bind.annotation.XmlAccessType;
import javax.xml.bind.annotation.XmlAccessorType;
import javax.xml.bind.annotation.XmlAttribute;
import javax.xml.bind.annotation.XmlElement;
import javax.xml.bind.annotation.XmlSchemaType;
import javax.xml.bind.annotation.XmlType;
import javax.xml.bind.annotation.XmlValue;`

	f.Write([]byte(fmt.Sprintf("%s\n\npackage %s;\n\n%s\n%s", copyright, packageName, importPackage, gen.Field)))
	return err
}

func genJavaFieldName(name string, unique bool) (fieldName string) {
	for _, str := range strings.Split(name, ":") {
		fieldName += MakeFirstUpperCase(str)
	}
	var tmp string
	for _, str := range strings.Split(fieldName, ".") {
		tmp += MakeFirstUpperCase(str)
	}
	fieldName = tmp
	fieldName = strings.Replace(fieldName, "-", "", -1)
	if unique {
		fieldNameCount[fieldName]++
		if count := fieldNameCount[fieldName]; count != 1 {
			fieldName = fmt.Sprintf("%s%d", fieldName, count)
		}
	}
	return
}

func genJavaFieldType(name string) string {
	if _, ok := javaBuildInType[name]; ok {
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

// JavaSimpleType generates code for simple type XML schema in Java language
// syntax.
func (gen *CodeGenerator) JavaSimpleType(v *SimpleType) {
	if v.List {
		if _, ok := gen.StructAST[v.Name]; !ok {
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			content := fmt.Sprintf("\tprotected List<%s> %s;\n", fieldType, genJavaFieldName(v.Name, false))
			gen.StructAST[v.Name] = content
			gen.Field += fmt.Sprintf("\n@XmlAccessorType(XmlAccessType.FIELD)\n@XmlAttribute(required = true, name = \"%s\")\npublic class %s {\n%s}\n", v.Name, genJavaFieldName(v.Name, true), gen.StructAST[v.Name])
			return
		}
	}
	if v.Union && len(v.MemberTypes) > 0 {
		if _, ok := gen.StructAST[v.Name]; !ok {
			content := " {\n"
			for _, member := range toSortedPairs(v.MemberTypes) {
				memberName := member.key
				memberType := member.value

				if memberType == "" { // fix order issue
					memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
				}
				fieldType := genJavaFieldType(memberType)
				content += fmt.Sprintf("\t@XmlElement(required = true)\n\tprotected %s %s;\n", fieldType, genJavaFieldName(memberName, false))
			}
			content += "}\n"
			gen.StructAST[v.Name] = content
			fieldName := genJavaFieldName(v.Name, true)
			gen.Field += fmt.Sprintf("%spublic class %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
		}
		return
	}
	if _, ok := gen.StructAST[v.Name]; !ok {
		fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
		content := fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(v.Name, false))
		gen.StructAST[v.Name] = content
		fieldName := genJavaFieldName(v.Name, true)
		gen.Field += fmt.Sprintf("%s@XmlAccessorType(XmlAccessType.FIELD)\n@XmlAttribute(required = true, name = \"%s\")\npublic class %s {\n%s}\n", genFieldComment(fieldName, v.Doc, "//"), v.Name, fieldName, gen.StructAST[v.Name])
	}
}

// JavaComplexType generates code for complex type XML schema in Java language
// syntax.
func (gen *CodeGenerator) JavaComplexType(v *ComplexType) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " {\n"
		for _, attrGroup := range v.AttributeGroup {
			fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
			content += fmt.Sprintf("\t@XmlElement(required = true)\n\tprotected %s %s;\n", genJavaFieldType(fieldType), genJavaFieldName(attrGroup.Name, false))
		}

		for _, attribute := range v.Attributes {
			required := ", required = true"
			if attribute.Optional {
				required = ""
			}
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
			content += fmt.Sprintf("\t@XmlAttribute(name = \"%s\"%s)\n\tprotected %s %sAttr;\n", attribute.Name, required, fieldType, genJavaFieldName(attribute.Name, false))
		}
		for _, group := range v.Groups {
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree))
			if group.Plural {
				fieldType = fmt.Sprintf("List<%s>", fieldType)
			}
			content += fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(group.Name, false))
		}

		for _, element := range v.Elements {
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
			if element.Plural {
				fieldType = fmt.Sprintf("List<%s>", fieldType)
			}
			content += fmt.Sprintf("\t@XmlElement(required = true, name = \"%s\")\n\tprotected %s %s;\n", element.Name, fieldType, genJavaFieldName(element.Name, false))
		}

		if len(v.Base) > 0 && isBuiltInJavaType(v.Base) {
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			content += fmt.Sprintf("\t@XmlValue\n\tprotected %s value;\n", fieldType)
		}

		content += "}\n"
		gen.StructAST[v.Name] = content
		fieldName := genJavaFieldName(v.Name, true)

		typeExtension := ""
		if len(v.Base) > 0 && !isBuiltInJavaType(v.Base) {
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			typeExtension = fmt.Sprintf(" extends %s ", fieldType)
		}

		gen.Field += fmt.Sprintf("%spublic class %s%s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, typeExtension, gen.StructAST[v.Name])
	}
}

func isBuiltInJavaType(typeName string) bool {
	_, builtIn := javaBuildInType[typeName]
	return builtIn
}

// JavaGroup generates code for group XML schema in Java language syntax.
func (gen *CodeGenerator) JavaGroup(v *Group) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " {\n"
		for _, element := range v.Elements {
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
			if element.Plural {
				fieldType = fmt.Sprintf("List<%s>", fieldType)
			}
			content += fmt.Sprintf("\t@XmlElement(required = true, name = \"%s\")\n\tprotected %s %s;\n", element.Name, fieldType, genJavaFieldName(element.Name, false))
		}

		for _, group := range v.Groups {
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree))
			if group.Plural {
				fieldType = fmt.Sprintf("List<%s>", fieldType)
			}
			content += fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(group.Name, false))
		}

		content += "}\n"
		gen.StructAST[v.Name] = content
		fieldName := genJavaFieldName(v.Name, true)
		gen.Field += fmt.Sprintf("%spublic class %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

// JavaAttributeGroup generates code for attribute group XML schema in Java language
// syntax.
func (gen *CodeGenerator) JavaAttributeGroup(v *AttributeGroup) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		content := " {\n"
		for _, attribute := range v.Attributes {
			required := ", required = true"
			if attribute.Optional {
				required = ""
			}
			fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
			content += fmt.Sprintf("\t@XmlAttribute(name = \"%s\"%s)\n\tprotected %sAttr %s;\n", attribute.Name, required, fieldType, genJavaFieldName(attribute.Name, false))
		}
		content += "}\n"
		gen.StructAST[v.Name] = content
		fieldName := genJavaFieldName(v.Name, true)
		gen.Field += fmt.Sprintf("%spublic class %s%s", genFieldComment(fieldName, v.Doc, "//"), fieldName, gen.StructAST[v.Name])
	}
}

// JavaElement generates code for element XML schema in Java language syntax.
func (gen *CodeGenerator) JavaElement(v *Element) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
		if v.Plural {
			fieldType = fmt.Sprintf("List<%s>", fieldType)
		}
		content := fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(v.Name, false))
		gen.StructAST[v.Name] = content
		gen.Field += fmt.Sprintf("\n@XmlAccessorType(XmlAccessType.FIELD)\n@XmlElement(required = true, name = \"%s\")\npublic class %s {\n%s}\n", v.Name, genJavaFieldName(v.Name, true), gen.StructAST[v.Name])
	}
}

// JavaAttribute generates code for attribute XML schema in Java language syntax.
func (gen *CodeGenerator) JavaAttribute(v *Attribute) {
	if _, ok := gen.StructAST[v.Name]; !ok {
		fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
		if v.Plural {
			fieldType = fmt.Sprintf("List<%s>", fieldType)
		}
		content := fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(v.Name, false))
		gen.StructAST[v.Name] = content
		gen.Field += fmt.Sprintf("\n@XmlAccessorType(XmlAccessType.FIELD)\n@XmlAttribute(required = true, name = \"%s\")\npublic class %s {\n%s}\n", v.Name, genJavaFieldName(v.Name, true), gen.StructAST[v.Name])
	}
}
