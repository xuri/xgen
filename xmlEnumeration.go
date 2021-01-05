// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

// OnEnumeration handles parsing event on the enumeration start elements.
func (opt *Options) OnEnumeration(ele xml.StartElement, protoTree []interface{}) (err error) {
	for _, attr := range ele.Attr {
		if attr.Name.Local == "value" {
			if opt.SimpleType.Peek() != nil {
				opt.SimpleType.Peek().(*SimpleType).Restriction.Enum = append(opt.SimpleType.Peek().(*SimpleType).Restriction.Enum, attr.Value)
			}
		}
	}
	return nil
}

// EndEnumeration handles parsing event on the enumeration end elements.
// Enumeration defines a list of acceptable values.
func (opt *Options) EndEnumeration(ele xml.EndElement, protoTree []interface{}) (err error) {
	if opt.Attribute.Len() > 0 && opt.SimpleType.Peek() != nil {
		if opt.Attribute.Peek().(*Attribute).Type, err = opt.GetValueType(opt.SimpleType.Peek().(*SimpleType).Base, opt.ProtoTree); err != nil {
			return
		}
		opt.CurrentEle = ""
	}
	if opt.SimpleType.Len() > 0 && opt.Element.Len() > 0 {
		if opt.Element.Peek().(*Element).Type, err = opt.GetValueType(opt.SimpleType.Peek().(*SimpleType).Base, opt.ProtoTree); err != nil {
			return
		}
		opt.CurrentEle = ""
	}
	return
}
