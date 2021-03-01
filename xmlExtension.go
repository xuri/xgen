// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

// OnExtension handles parsing event on the extension start elements. The
// extension element defines a base class for a complexType or simpleContent.
func (opt *Options) OnExtension(ele xml.StartElement, protoTree []interface{}) (err error) {
	for _, attr := range ele.Attr {
		if attr.Name.Local == "base" {
			var valueType string
			valueType, err = opt.GetValueType(attr.Value, protoTree)
			if err != nil {
				return
			}
			if opt.ComplexType.Peek() != nil {
				var complexType = opt.ComplexType.Peek().(*ComplexType)
				complexType.Base, err = opt.GetValueType(valueType, protoTree)
				if err != nil {
					return
				}
				if complexType.Name == "" {
					complexType.Name = attr.Value
				}
			}
		}
	}
	return
}

// EndExtension handles parsing event on the extension end elements.
func (opt *Options) EndExtension(ele xml.EndElement, protoTree []interface{}) (err error) {
	if opt.Attribute.Len() > 0 && opt.SimpleType.Peek() != nil {
		opt.Attribute.Peek().(*Attribute).Type, err = opt.GetValueType(opt.SimpleType.Pop().(*SimpleType).Base, opt.ProtoTree)
		if err != nil {
			return
		}
		opt.CurrentEle = ""
	}
	return
}
