// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"encoding/xml"
	"strconv"
)

// OnElement handles parsing event on the element start elements.
func (opt *Options) OnElement(ele xml.StartElement, protoTree []interface{}) (err error) {
	e := Element{}
	for _, attr := range ele.Attr {
		if attr.Name.Local == "ref" {
			e.Name = attr.Value
			e.Type, err = opt.GetValueType(attr.Value, protoTree)
			if err != nil {
				return
			}
		}

		if attr.Name.Local == "name" {
			e.Name = attr.Value
		}
		if attr.Name.Local == "type" {
			e.Type, err = opt.GetValueType(attr.Value, protoTree)
			if err != nil {
				return
			}
		}
		if attr.Name.Local == "maxOccurs" {
			var maxOccurs int
			if maxOccurs, err = strconv.Atoi(attr.Value); attr.Value != "unbounded" && err != nil {
				return
			}
			if attr.Value == "unbounded" || maxOccurs > 1 {
				e.Plural, err = true, nil
			}
		}
		if attr.Name.Local == "unbounded" {
			if attr.Value != "0" {
				e.Plural = true
			}
		}
	}

	if e.Type == "" {
		e.Type, err = opt.GetValueType(e.Name, protoTree)
		if err != nil {
			return
		}
		opt.Element.Push(&e)
	}
	if opt.ComplexType.Len() > 0 {
		if !inElements(&e, opt.ComplexType.Peek().(*ComplexType).Elements) {
			opt.ComplexType.Peek().(*ComplexType).Elements = append(opt.ComplexType.Peek().(*ComplexType).Elements, e)
		}
		return
	}

	if opt.InGroup > 0 {
		if opt.Group.Len() > 0 {
			opt.Group.Peek().(*Group).Elements = append(opt.Group.Peek().(*Group).Elements, e)
		}
		return
	}

	opt.Element.Push(&e)
	return
}

// EndElement handles parsing event on the element end elements.
func (opt *Options) EndElement(ele xml.EndElement, protoTree []interface{}) (err error) {
	if opt.Element.Len() > 0 && opt.ComplexType.Len() == 0 {
		opt.ProtoTree = append(opt.ProtoTree, opt.Element.Pop())
	}
	return
}

func inElements(element *Element, elements []Element) bool {
	for _, ele := range elements {
		if element.Name == ele.Name {
			return true
		}
	}
	return false
}
