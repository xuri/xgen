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

var TypeScriptBuildInType = map[string]bool{
	"boolean":   true,
	"number":    true,
	"string":    true,
	"void":      true,
	"null":      true,
	"undefined": true,
}

// GenTypeScript generate TypeScript programming language source code for XML
// schema definition files.
func (gen *CodeGenerator) GenTypeScript() error {
	structAST := map[string]string{}
	var field string
	for _, ele := range gen.ProtoTree {
		switch v := ele.(type) {
		case *SimpleType:
			if v.List {
				if _, ok := structAST[v.Name]; !ok {
					fieldType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
					content := fmt.Sprintf(" Array<%s>\n", genTypeScriptFieldType(fieldType))
					structAST[v.Name] = content
					field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
					continue
				}
			}
			if v.Union && len(v.MemberTypes) > 0 {
				if _, ok := structAST[v.Name]; !ok {
					content := " {\n"
					for memberName, memberType := range v.MemberTypes {
						if memberType == "" { // fix order issue
							memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
						}
						content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(memberName), genTypeScriptFieldType(memberType))
					}
					content += "}\n"
					structAST[v.Name] = content
					field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
				}
				continue
			}
			if _, ok := structAST[v.Name]; !ok {
				content := fmt.Sprintf(" %s;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree)))
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport type %s =%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
			}

		case *ComplexType:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, attrGroup := range v.AttributeGroup {
					fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(attrGroup.Name), genTypeScriptFieldType(fieldType))
				}

				for _, attribute := range v.Attributes {
					var optional string
					if attribute.Optional {
						optional = ` | null`
					}
					fieldType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
					content += fmt.Sprintf("\t%sAttr: %s%s;\n", genTypeScriptFieldName(attribute.Name), fieldType, optional)
				}
				for _, group := range v.Groups {
					if group.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFieldName(group.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
						continue
					}
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(group.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
				}

				for _, element := range v.Elements {
					fieldType := genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
					if element.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFieldName(element.Name), fieldType)
						continue
					}
					content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFieldName(element.Name), fieldType)
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
			}

		case *Group:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, element := range v.Elements {
					if element.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFieldName(element.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)))
						continue
					}
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(element.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)))
				}

				for _, group := range v.Groups {
					if group.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFieldName(group.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
						continue
					}
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFieldName(group.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
				}

				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
			}

		case *AttributeGroup:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, attribute := range v.Attributes {
					var optional string
					if attribute.Optional {
						optional = ` | null`
					}
					content += fmt.Sprintf("\t%sAttr: %s%s;\n", genTypeScriptFieldName(attribute.Name), genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)), optional)
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
			}

		case *Element:
			if _, ok := structAST[v.Name]; !ok {
				if v.Plural {
					structAST[v.Name] = fmt.Sprintf(" Array<%s>;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				} else {
					structAST[v.Name] = fmt.Sprintf(" %s;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				}

				field += fmt.Sprintf("\nexport type %s =%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
			}

		case *Attribute:
			if _, ok := structAST[v.Name]; !ok {
				if v.Plural {
					structAST[v.Name] = fmt.Sprintf(" Array<%s>;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				} else {
					structAST[v.Name] = fmt.Sprintf(" %s;\n", genTypeScriptFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				}
				field += fmt.Sprintf("\nexport type %s =%s", genTypeScriptFieldName(v.Name), structAST[v.Name])
			}
		}
	}
	f, err := os.Create(gen.File + ".ts")
	if err != nil {
		return err
	}
	defer f.Close()
	source := []byte(fmt.Sprintf("%s\n%s", copyright, field))
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

func genTypeScriptFieldType(name string) string {
	if _, ok := TypeScriptBuildInType[name]; ok {
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
	return "any"
}
