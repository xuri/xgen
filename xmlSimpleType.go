// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

func (opt *Options) OnSimpleType(ele xml.StartElement, protoTree []interface{}) (err error) {
	if opt.SimpleType.Len() == 0 {
		opt.SimpleType.Push(&SimpleType{})
	}
	opt.CurrentEle = opt.InElement
	for _, attr := range ele.Attr {
		if attr.Name.Local == "name" {
			opt.SimpleType.Peek().(*SimpleType).Name = attr.Value
		}
	}
	return
}
