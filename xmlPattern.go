// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

// EndPattern handles parsing event on the pattern end elements. Pattern
// defines the exact sequence of characters that are acceptable.
func (opt *Options) EndPattern(ele xml.EndElement, protoTree []interface{}) (err error) {
	if opt.Attribute.Len() > 0 && opt.SimpleType.Peek() != nil {
		opt.Attribute.Peek().(*Attribute).Type, err = opt.GetValueType(opt.SimpleType.Pop().(*SimpleType).Base, opt.ProtoTree)
		if err != nil {
			return
		}
		opt.CurrentEle = ""
	}
	if opt.SimpleType.Len() > 0 && opt.Element.Len() > 0 {
		if opt.Element.Peek().(*Element).Type, err = opt.GetValueType(opt.SimpleType.Pop().(*SimpleType).Base, opt.ProtoTree); err != nil {
			return
		}
		opt.CurrentEle = ""
	}
	return
}
