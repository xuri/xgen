// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"encoding/xml"
	"strings"
)

func (opt *Options) OnUnion(ele xml.StartElement, protoTree []interface{}) (err error) {
	opt.InUnion = true
	if opt.SimpleType.Peek() == nil {
		return
	}
	opt.SimpleType.Peek().(*SimpleType).Union = true
	opt.SimpleType.Peek().(*SimpleType).MemberTypes = make(map[string]string)
	for _, attr := range ele.Attr {
		if attr.Name.Local == "memberTypes" {
			memberTypes := strings.Split(attr.Value, " ")
			for _, memberType := range memberTypes {
				opt.SimpleType.Peek().(*SimpleType).MemberTypes[trimNSPrefix(memberType)], err = opt.GetValueType(memberType, protoTree)
				if err != nil {
					return
				}
			}
			continue
		}
	}
	return
}
