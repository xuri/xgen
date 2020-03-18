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

// GenC generate C programming language source code for XML schema definition
// files.
func (gen *CodeGenerator) GenC() error {
	structAST := map[string]string{}
	var field string
	for _, ele := range gen.ProtoTree {
		switch v := ele.(type) {
		case *SimpleType:
			if v.List {
				if _, ok := structAST[v.Name]; !ok {
					fieldType := genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
					content := fmt.Sprintf("%s %s[];\n", genCFieldType(fieldType), genCFieldName(v.Name))
					structAST[v.Name] = content
					field += fmt.Sprintf("\ntypedef %s", structAST[v.Name])
					continue
				}
			}
			if v.Union && len(v.MemberTypes) > 0 {
				if _, ok := structAST[v.Name]; !ok {
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
					structAST[v.Name] = content
					field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFieldName(v.Name))
				}
				continue
			}
			if _, ok := structAST[v.Name]; !ok {
				var plural, fieldType string
				var ok bool
				if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))); ok {
					plural = "[]"
				}
				structAST[v.Name] = fmt.Sprintf("%s %s%s", fieldType, genCFieldName(v.Name), plural)
				field += fmt.Sprintf("\ntypedef %s;\n", structAST[v.Name])
			}
		case *ComplexType:
			if _, ok := structAST[v.Name]; !ok {
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
				content += "}"
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFieldName(v.Name))
			}
		case *Group:
			if _, ok := structAST[v.Name]; !ok {
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
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFieldName(v.Name))
			}
		case *AttributeGroup:
			if _, ok := structAST[v.Name]; !ok {
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
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFieldName(v.Name))
			}
		case *Element:
			if _, ok := structAST[v.Name]; !ok {
				var plural, fieldType string
				var ok bool
				if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))); ok || v.Plural {
					plural = "[]"
				}
				structAST[v.Name] = fmt.Sprintf("%s %s%s", fieldType, genCFieldName(v.Name), plural)
				field += fmt.Sprintf("\ntypedef %s;\n", structAST[v.Name])
			}
		case *Attribute:
			if _, ok := structAST[v.Name]; !ok {
				var plural, fieldType string
				var ok bool
				if fieldType, ok = innerArray(genCFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))); ok || v.Plural {
					plural = "[]"
				}
				structAST[v.Name] = fmt.Sprintf("%s %s%s", fieldType, genCFieldName(v.Name), plural)
				field += fmt.Sprintf("\ntypedef %s;\n", structAST[v.Name])
			}
		}
	}
	f, err := os.Create(gen.File + ".h")
	if err != nil {
		return err
	}
	defer f.Close()
	source := []byte(fmt.Sprintf("%s\n%s", copyright, field))
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
