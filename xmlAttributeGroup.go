// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

func (opt *Options) OnAttributeGroup(ele xml.StartElement, protoTree []interface{}) (err error) {
	attributeGroup := AttributeGroup{}
	for _, attr := range ele.Attr {
		if attr.Name.Local == "name" {
			attributeGroup.Name = attr.Value
		}
		if attr.Name.Local == "ref" {
			attributeGroup.Name = attr.Value
			attributeGroup.Ref, err = opt.GetValueType(attr.Value, protoTree)
			if err != nil {
				return
			}
		}
	}
	if opt.ComplexType.Len() == 0 {
		opt.InAttributeGroup = true
		opt.CurrentEle = opt.InElement
		opt.AttributeGroup = &attributeGroup
		return
	}

	if opt.ComplexType.Len() > 0 {
		opt.ComplexType.Peek().(*ComplexType).AttributeGroup = append(opt.ComplexType.Peek().(*ComplexType).AttributeGroup, attributeGroup)
		return
	}
	return
}
