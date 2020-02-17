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
					filedType := genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
					content := fmt.Sprintf(" Array<%s>\n", genTypeScriptFiledType(filedType))
					structAST[v.Name] = content
					field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
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
						content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFiledName(memberName), genTypeScriptFiledType(memberType))
					}
					content += "}\n"
					structAST[v.Name] = content
					field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
				}
				continue
			}
			if _, ok := structAST[v.Name]; !ok {
				content := fmt.Sprintf(" %s;\n", genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree)))
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport type %s =%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
			}

		case *ComplexType:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, attrGroup := range v.AttributeGroup {
					filedType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFiledName(attrGroup.Name), genTypeScriptFiledType(filedType))
				}

				for _, attribute := range v.Attributes {
					var optional string
					if attribute.Optional {
						optional = ` | null`
					}
					filedType := genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
					content += fmt.Sprintf("\t%sAttr: %s%s;\n", genTypeScriptFiledName(attribute.Name), filedType, optional)
				}
				for _, group := range v.Groups {
					if group.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFiledName(group.Name), genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
						continue
					}
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFiledName(group.Name), genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
				}

				for _, element := range v.Elements {
					filedType := genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
					if element.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFiledName(element.Name), filedType)
						continue
					}
					content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFiledName(element.Name), filedType)
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
			}

		case *Group:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, element := range v.Elements {
					if element.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFiledName(element.Name), genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)))
						continue
					}
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFiledName(element.Name), genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)))
				}

				for _, group := range v.Groups {
					if group.Plural {
						content += fmt.Sprintf("\t%s: Array<%s>;\n", genTypeScriptFiledName(group.Name), genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
						continue
					}
					content += fmt.Sprintf("\t%s: %s;\n", genTypeScriptFiledName(group.Name), genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
				}

				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
			}

		case *AttributeGroup:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, attribute := range v.Attributes {
					var optional string
					if attribute.Optional {
						optional = ` | null`
					}
					content += fmt.Sprintf("\t%sAttr: %s%s;\n", genTypeScriptFiledName(attribute.Name), genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)), optional)
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\nexport class %s%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
			}

		case *Element:
			if _, ok := structAST[v.Name]; !ok {
				if v.Plural {
					structAST[v.Name] = fmt.Sprintf(" Array<%s>;\n", genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				} else {
					structAST[v.Name] = fmt.Sprintf(" %s;\n", genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				}

				field += fmt.Sprintf("\nexport type %s =%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
			}

		case *Attribute:
			if _, ok := structAST[v.Name]; !ok {
				if v.Plural {
					structAST[v.Name] = fmt.Sprintf(" Array<%s>;\n", genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				} else {
					structAST[v.Name] = fmt.Sprintf(" %s;\n", genTypeScriptFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				}
				field += fmt.Sprintf("\nexport type %s =%s", genTypeScriptFiledName(v.Name), structAST[v.Name])
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

func genTypeScriptFiledName(name string) (filedName string) {
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

func genTypeScriptFiledType(name string) string {
	if _, ok := TypeScriptBuildInType[name]; ok {
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
	return "any"
}
