// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

// OnRestriction handles parsing event on the restriction start elements. The
// restriction element defines restrictions on a simpleType, simpleContent, or
// complexContent definition.
func (opt *Options) OnRestriction(ele xml.StartElement, protoTree []interface{}) (err error) {
	for _, attr := range ele.Attr {
		if attr.Name.Local == "base" {
			var valueType string
			valueType, err = opt.GetValueType(attr.Value, protoTree)
			if err != nil {
				return
			}
			// If inside a simple type definition, we cannot also be inside a simpleContent or complexContent
			if opt.SimpleType.Peek() != nil {
				opt.SimpleType.Peek().(*SimpleType).Base, err = opt.GetValueType(valueType, protoTree)
				if err != nil {
					return
				}
				if opt.SimpleType.Peek().(*SimpleType).Name == "" {
					opt.SimpleType.Peek().(*SimpleType).Name = attr.Value
				}
			}
		}
	}
	return
}

// EndRestriction handles parsing event on the restriction end elements.
func (opt *Options) EndRestriction(ele xml.EndElement, protoTree []interface{}) (err error) {
	if opt.Attribute.Len() > 0 && opt.SimpleType.Peek() != nil {
		opt.Attribute.Peek().(*Attribute).Type, err = opt.GetValueType(opt.SimpleType.Pop().(*SimpleType).Base, opt.ProtoTree)
		if err != nil {
			return
		}
		opt.CurrentEle = ""
	}
	if !opt.Element.Empty() {
		// This handles the case of a simpleContent or complexContent (XXX - is this correct?)
		if !opt.ComplexType.Empty() {
			var complexType = opt.ComplexType.Peek().(*ComplexType)
			if len(complexType.Elements) > 0 {
				complexType.Elements = append(complexType.Elements, *opt.Element.Peek().(*Element))
			}
		}
	}
	return
}
