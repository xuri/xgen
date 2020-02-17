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
	"go/format"
	"os"
	"strings"
)

// CodeGenerator holds code generator overrides and runtime data that are used
// when generate code from proto tree.
type CodeGenerator struct {
	Lang      string
	File      string
	ProtoTree []interface{}
}

var goBuildinType = map[string]bool{
	"string":    true,
	"[]string":  true,
	"xml.Name":  true,
	"[]byte":    true,
	"bool":      true,
	"byte":      true,
	"time.Time": true,
	"float64":   true,
	"float32":   true,
	"int":       true,
	"int64":     true,
	"uint":      true,
	"uint64":    true,
}

// GenTypeScript generate Go programming language source code for XML schema
// definition files.
func (gen *CodeGenerator) GenGo() error {
	structAST := map[string]string{}
	var field string
	var importTime bool
	for _, ele := range gen.ProtoTree {
		switch v := ele.(type) {
		case *SimpleType:
			if v.List {
				if _, ok := structAST[v.Name]; !ok {
					filedType := genFiledType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
					if filedType == "time.Time" {
						importTime = true
					}
					content := fmt.Sprintf(" []%s\n", genFiledType(filedType))
					structAST[v.Name] = content
					field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])
					continue
				}
			}
			if v.Union && len(v.MemberTypes) > 0 {
				if _, ok := structAST[v.Name]; !ok {
					content := " struct {\n"
					for memberName, memberType := range v.MemberTypes {
						if memberType == "" { // fix order issue
							memberType = getBasefromSimpleType(memberName, gen.ProtoTree)
						}
						content += fmt.Sprintf("\t%s\t%s\n", genFiledName(memberName), genFiledType(memberType))
					}
					content += "}\n"
					structAST[v.Name] = content
					field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])

				}
				continue
			}
			if _, ok := structAST[v.Name]; !ok {
				content := fmt.Sprintf(" %s\n", genFiledType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree)))
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])
			}

		case *ComplexType:
			if _, ok := structAST[v.Name]; !ok {
				content := " struct {\n"
				for _, attrGroup := range v.AttributeGroup {
					filedType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
					if filedType == "time.Time" {
						importTime = true
					}
					content += fmt.Sprintf("\t%s\t%s\n", genFiledName(attrGroup.Name), genFiledType(filedType))
				}

				for _, attribute := range v.Attributes {
					var optional string
					if attribute.Optional {
						optional = `,omitempty`
					}
					filedType := genFiledType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
					if filedType == "time.Time" {
						importTime = true
					}
					content += fmt.Sprintf("\t%sAttr\t%s\t`xml:\"%s,attr%s\"`\n", genFiledName(attribute.Name), filedType, attribute.Name, optional)
				}
				for _, group := range v.Groups {
					var plural string
					if group.Plural {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s\t%s%s\n", genFiledName(group.Name), plural, genFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
				}

				for _, element := range v.Elements {
					var plural string
					if element.Plural {
						plural = "[]"
					}
					filedType := genFiledType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
					if filedType == "time.Time" {
						importTime = true
					}
					content += fmt.Sprintf("\t%s\t%s%s\t`xml:\"%s\"`\n", genFiledName(element.Name), plural, filedType, element.Name)
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])
			}

		case *Group:
			if _, ok := structAST[v.Name]; !ok {
				content := " struct {\n"
				for _, element := range v.Elements {
					var plural string
					if element.Plural {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s\t%s%s\n", genFiledName(element.Name), plural, genFiledType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)))
				}

				for _, group := range v.Groups {
					var plural string
					if group.Plural {
						plural = "[]"
					}
					content += fmt.Sprintf("\t%s\t%s%s\n", genFiledName(group.Name), plural, genFiledType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)))
				}

				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])
			}
		case *AttributeGroup:
			if _, ok := structAST[v.Name]; !ok {
				content := " struct {\n"
				for _, attribute := range v.Attributes {
					var optional string
					if attribute.Optional {
						optional = `,omitempty`
					}
					content += fmt.Sprintf("\t%sAttr\t%s\t`xml:\"%s,attr%s\"`\n", genFiledName(attribute.Name), genFiledType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)), attribute.Name, optional)
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])

			}
		case *Element:
			if _, ok := structAST[v.Name]; !ok {
				var plural string
				if v.Plural {
					plural = "[]"
				}
				content := fmt.Sprintf("\t%s%s\n", plural, genFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])
			}

		case *Attribute:
			if _, ok := structAST[v.Name]; !ok {
				var plural string
				if v.Plural {
					plural = "[]"
				}
				content := fmt.Sprintf("\t%s%s\n", plural, genFiledType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)))
				structAST[v.Name] = content
				field += fmt.Sprintf("\ntype %s%s", genFiledName(v.Name), structAST[v.Name])
			}
		}
	}
	f, err := os.Create(gen.File + ".go")
	if err != nil {
		return err
	}
	defer f.Close()
	var importPackage string
	if importTime {
		importPackage = "import (\n\t\"time\"\n)"
	}
	source, err := format.Source([]byte(fmt.Sprintf("%s\n\npackage schema\n%s%s", copyright, importPackage, field)))
	if err != nil {
		f.WriteString(fmt.Sprintf("package schema\n%s%s", importPackage, field))
		return err
	}
	f.Write(source)
	return err
}

func genFiledName(name string) (filedName string) {
	for _, str := range strings.Split(name, ":") {
		filedName += MakeFirstUpperCase(str)
	}
	var tmp string
	for _, str := range strings.Split(filedName, ".") {
		tmp += MakeFirstUpperCase(str)
	}
	filedName = tmp
	filedName = strings.Replace(strings.Replace(filedName, "-", "", -1), "_", "", -1)
	return
}

func genFiledType(name string) string {
	if _, ok := goBuildinType[name]; ok {
		return name
	}
	var filedType string
	for _, str := range strings.Split(name, ".") {
		filedType += MakeFirstUpperCase(str)
	}
	filedType = strings.Replace(MakeFirstUpperCase(strings.Replace(filedType, "-", "", -1)), "_", "", -1)
	if filedType != "" {
		return "*" + filedType
	}
	return "interface{}"
}

var copyright = `// Copyright 2020 The xgen Authors. All rights reserved.
//
// DO NOT EDIT: generated by xgen XSD generator
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.`
