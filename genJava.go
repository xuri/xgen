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

var JavaBuildInType = map[string]bool{
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
	structAST := map[string]string{}
	var field string
	for _, ele := range gen.ProtoTree {
		switch v := ele.(type) {
		case *SimpleType:
			if v.List {
				if _, ok := structAST[v.Name]; !ok {
					fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
					content := fmt.Sprintf("\tprotected List<%s> %s;\n", fieldType, genJavaFieldName(v.Name))
					structAST[v.Name] = content
					field += fmt.Sprintf("\n@XmlAccessorType(XmlAccessType.FIELD)\n@XmlAttribute(required = true, name = \"%s\")\npublic class %s {\n%s}\n", v.Name, genJavaFieldName(v.Name), structAST[v.Name])
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
						fieldType := genJavaFieldType(memberType)
						content += fmt.Sprintf("\t@XmlElement(required = true)\n\tprotected %s %s;\n", fieldType, genJavaFieldName(memberName))
					}
					content += "}\n"
					structAST[v.Name] = content
					field += fmt.Sprintf("\npublic class %s%s", genJavaFieldName(v.Name), structAST[v.Name])
				}
				continue
			}
			if _, ok := structAST[v.Name]; !ok {
				fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
				content := fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(v.Name))
				structAST[v.Name] = content
				field += fmt.Sprintf("\n@XmlAccessorType(XmlAccessType.FIELD)\n@XmlAttribute(required = true, name = \"%s\")\npublic class %s {\n%s}\n", v.Name, genJavaFieldName(v.Name), structAST[v.Name])
			}

		case *ComplexType:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, attrGroup := range v.AttributeGroup {
					fieldType := getBasefromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
					content += fmt.Sprintf("\t@XmlElement(required = true)\n\tprotected %s %s;\n", genJavaFieldType(fieldType), genJavaFieldName(attrGroup.Name))
				}

				for _, attribute := range v.Attributes {
					var required = ", required = true"
					if attribute.Optional {
						required = ""
					}
					fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
					content += fmt.Sprintf("\t@XmlAttribute(name = \"%s\"%s)\n\tprotected %sAttr %s;\n", attribute.Name, required, fieldType, genJavaFieldName(attribute.Name))
				}
				for _, group := range v.Groups {
					var fieldType = genJavaFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree))
					if group.Plural {
						fieldType = fmt.Sprintf("List<%s>", fieldType)
					}
					content += fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(group.Name))
				}

				for _, element := range v.Elements {
					fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
					if element.Plural {
						fieldType = fmt.Sprintf("List<%s>", fieldType)
					}
					content += fmt.Sprintf("\t@XmlElement(required = true, name = \"%s\")\n\tprotected %s %s;\n", element.Name, fieldType, genJavaFieldName(element.Name))
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\npublic class %s%s", genJavaFieldName(v.Name), structAST[v.Name])
			}

		case *Group:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, element := range v.Elements {
					var fieldType = genJavaFieldType(getBasefromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
					if element.Plural {
						fieldType = fmt.Sprintf("List<%s>", fieldType)
					}
					content += fmt.Sprintf("\t@XmlElement(required = true, name = \"%s\")\n\tprotected %s %s;\n", element.Name, fieldType, genJavaFieldName(element.Name))
				}

				for _, group := range v.Groups {
					var fieldType = genJavaFieldType(getBasefromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree))
					if group.Plural {
						fieldType = fmt.Sprintf("List<%s>", fieldType)
					}
					content += fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(group.Name))
				}

				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\npublic class %s%s", genJavaFieldName(v.Name), structAST[v.Name])
			}
		case *AttributeGroup:
			if _, ok := structAST[v.Name]; !ok {
				content := " {\n"
				for _, attribute := range v.Attributes {
					var required = ", required = true"
					if attribute.Optional {
						required = ""
					}
					fieldType := genJavaFieldType(getBasefromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
					content += fmt.Sprintf("\t@XmlAttribute(name = \"%s\"%s)\n\tprotected %sAttr %s;\n", attribute.Name, required, fieldType, genJavaFieldName(attribute.Name))
				}
				content += "}\n"
				structAST[v.Name] = content
				field += fmt.Sprintf("\npublic class %s%s", genJavaFieldName(v.Name), structAST[v.Name])

			}
		case *Element:
			if _, ok := structAST[v.Name]; !ok {
				var fieldType = genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
				if v.Plural {
					fieldType = fmt.Sprintf("List<%s>", fieldType)
				}
				content := fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(v.Name))
				structAST[v.Name] = content
				field += fmt.Sprintf("\n@XmlAccessorType(XmlAccessType.FIELD)\n@XmlElement(required = true, name = \"%s\")\npublic class %s {\n%s}\n", v.Name, genJavaFieldName(v.Name), structAST[v.Name])
			}

		case *Attribute:
			if _, ok := structAST[v.Name]; !ok {
				var fieldType = genJavaFieldType(getBasefromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree))
				if v.Plural {
					fieldType = fmt.Sprintf("List<%s>", fieldType)
				}
				content := fmt.Sprintf("\tprotected %s %s;\n", fieldType, genJavaFieldName(v.Name))
				structAST[v.Name] = content
				field += fmt.Sprintf("\n@XmlAccessorType(XmlAccessType.FIELD)\n@XmlAttribute(required = true, name = \"%s\")\npublic class %s {\n%s}\n", v.Name, genJavaFieldName(v.Name), structAST[v.Name])
			}
		}
	}
	f, err := os.Create(gen.File + ".java")
	if err != nil {
		return err
	}
	defer f.Close()
	var importPackage = `import java.util.ArrayList;
import java.util.List;
import javax.xml.bind.annotation.XmlAccessType;
import javax.xml.bind.annotation.XmlAccessorType;
import javax.xml.bind.annotation.XmlAttribute;
import javax.xml.bind.annotation.XmlElement;
import javax.xml.bind.annotation.XmlSchemaType;
import javax.xml.bind.annotation.XmlType;`

	f.Write([]byte(fmt.Sprintf("%s\n\npackage schema;\n\n%s\n%s", copyright, importPackage, field)))
	return err
}

func genJavaFieldName(name string) (fieldName string) {
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

func genJavaFieldType(name string) string {
	if _, ok := JavaBuildInType[name]; ok {
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
