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

var CBuildInType = map[string]bool{
	"bool":           true,
	"char":           true,
	"unsigned char":  true,
	"signed char":    true,
	"char[]":         true, // char[] will be flat to 'char filed_name[]'
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
					filedType := genCFiledType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
					content := fmt.Sprintf("%s %s[];\n", genCFiledType(filedType), genCFiledName(v.Name))
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
						var plural, filedType string
						var ok bool
						if filedType, ok = innerArray(genCFiledType(memberType)); ok {
							plural = "[]"
						}
						content += fmt.Sprintf("\t%s %s%s;\n", filedType, genCFiledName(memberName), plural)
					}
					content += "}"
					structAST[v.Name] = content
					field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFiledName(v.Name))
				}
				continue
			}
			if _, ok := structAST[v.Name]; !ok {
				var plural, filedType string
				var ok bool
				if filedType, ok = innerArray(genCFiledType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))); ok {
					plural = "[]"
				}
				structAST[v.Name] = fmt.Sprintf("%s %s%s", filedType, genCFiledName(v.Name), plural)
				field += fmt.Sprintf("\ntypedef %s;\n", structAST[v.Name])
			}
		case *ComplexType:
			if _, ok := structAST[v.Name]; !ok {
				content := "struct {\n"
				for _, attrGroup := range v.AttributeGroup {
					filedType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
					content += fmt.Sprintf("\t%s %s;\n", genCFiledType(filedType), genCFiledName(attrGroup.Name))
				}

				for _, attribute := range v.Attributes {
					var optional string
					if attribute.Optional {
						optional = `, optional`
					}
					var plural, filedType string
					var ok bool
					if filedType, ok = innerArray(genCFiledType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))); ok {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s %sAttr%s; // attr%s\n", filedType, genCFiledName(attribute.Name), plural, optional)
				}

				for _, group := range v.Groups {
					var plural string
					if group.Plural {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s %s%s;\n", genCFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)), genCFiledName(group.Name), plural)
				}

				for _, element := range v.Elements {
					var plural, filedType string
					var ok bool
					if filedType, ok = innerArray(genCFiledType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))); ok || element.Plural {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s %s%s;\n", filedType, genCFiledName(element.Name), plural)
				}
				content += "}"
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFiledName(v.Name))
			}
		case *Group:
			if _, ok := structAST[v.Name]; !ok {
				content := "struct {\n"
				for _, element := range v.Elements {
					var plural string
					if element.Plural {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s %s%s;\n", genCFiledType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)), genCFiledName(element.Name), plural)
				}

				for _, group := range v.Groups {
					var plural string
					if group.Plural {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s %s%s;\n", genCFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)), genCFiledName(group.Name), plural)
				}

				content += "}"
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFiledName(v.Name))
			}
		case *AttributeGroup:
			if _, ok := structAST[v.Name]; !ok {
				content := "struct {\n"
				for _, attribute := range v.Attributes {
					var optional, plural, filedType string
					var ok bool
					if attribute.Optional {
						optional = `, optional`
					}
					if filedType, ok = innerArray(genCFiledType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))); ok {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s %sAttr%s; // attr%s\n", filedType, genCFiledName(attribute.Name), plural, optional)
				}
				content += "}"
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntypedef %s %s;\n", structAST[v.Name], genCFiledName(v.Name))
			}
		case *Element:
			if _, ok := structAST[v.Name]; !ok {
				var plural, filedType string
				var ok bool
				if filedType, ok = innerArray(genCFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))); ok || v.Plural {
					plural = "[]"
				}
				structAST[v.Name] = fmt.Sprintf("%s %s%s", filedType, genCFiledName(v.Name), plural)
				field += fmt.Sprintf("\ntypedef %s;\n", structAST[v.Name])
			}
		case *Attribute:
			if _, ok := structAST[v.Name]; !ok {
				var plural, filedType string
				var ok bool
				if filedType, ok = innerArray(genCFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))); ok || v.Plural {
					plural = "[]"
				}
				structAST[v.Name] = fmt.Sprintf("%s %s%s", filedType, genCFiledName(v.Name), plural)
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

func genCFiledName(name string) (filedName string) {
	for _, str := range strings.Split(name, ":") {
		filedName += MakeFirstUpperCase(str)
	}
	var tmp string
	for _, str := range strings.Split(filedName, ".") {
		tmp += MakeFirstUpperCase(str)
	}
	filedName = tmp
	filedName = strings.Replace(filedName, "-", "", -1)
	return
}

func genCFiledType(name string) string {
	if _, ok := CBuildInType[name]; ok {
		return name
	}
	var filedType string
	for _, str := range strings.Split(name, ".") {
		filedType += MakeFirstUpperCase(str)
	}
	filedType = MakeFirstUpperCase(strings.Replace(filedType, "-", "", -1))
	if filedType != "" {
		return filedType
	}
	return "void"
}
