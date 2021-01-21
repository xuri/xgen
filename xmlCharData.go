// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "strings"

// OnCharData handles parsing event on the documentation start elements. The
// documentation element specifies information to be read or used by users
// within an annotation element.
func (opt *Options) OnCharData(ele string, protoTree []interface{}) (err error) {
	if strings.TrimSpace(ele) == "" {
		return
	}
	ele = strings.TrimSpace(ele)
	if opt.InAttributeGroup {
		if opt.AttributeGroup.Peek() != nil {
			opt.AttributeGroup.Peek().(*AttributeGroup).Doc = ele
			return
		}
	}
	if opt.InElement != "" {
		if opt.Element.Peek() != nil {
			opt.Element.Peek().(*Element).Doc = ele
			return
		}
	}
	if opt.Attribute.Len() > 0 {
		opt.Attribute.Peek().(*Attribute).Doc = ele
		return
	}
	switch opt.CurrentEle {
	case "simpleType":
		if opt.SimpleType.Peek() != nil {
			opt.SimpleType.Peek().(*SimpleType).Doc = ele
			return
		}
	case "complexType":
		if opt.Attribute.Len() > 0 {
			opt.Attribute.Peek().(*Attribute).Doc = ele
			return
		}
		if opt.ComplexType.Peek() != nil {
			l := len(opt.ComplexType.Peek().(*ComplexType).Attributes)
			if l > 0 {
				opt.ComplexType.Peek().(*ComplexType).Attributes[l-1].Doc = ele
				return
			}
			opt.ComplexType.Peek().(*ComplexType).Doc = ele
			return
		}
	default:
	}
	return
}
