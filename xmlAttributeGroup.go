// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

// OnAttributeGroup handles parsing event on the attributeGroup start
// elements. The attributeGroup element is used to group a set of attribute
// declarations so that they can be incorporated as a group into complex type
// definitions.
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
		opt.AttributeGroup.Push(&attributeGroup)
		return
	}

	if opt.ComplexType.Len() > 0 {
		opt.ComplexType.Peek().(*ComplexType).AttributeGroup = append(opt.ComplexType.Peek().(*ComplexType).AttributeGroup, attributeGroup)
		return
	}
	return
}

// EndAttributeGroup handles parsing event on the attributeGroup end elements.
func (opt *Options) EndAttributeGroup(ele xml.EndElement, protoTree []interface{}) (err error) {
	if opt.AttributeGroup.Len() > 0 {
		opt.ProtoTree = append(opt.ProtoTree, opt.AttributeGroup.Pop())
		opt.CurrentEle = ""
		opt.InAttributeGroup = false
	}
	return
}
