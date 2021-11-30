// Copyright 2021 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"encoding/xml"
	"strconv"
)

// OnChoice handles parsing event on the choice start elements. The
// choice element defines that one and only one of the contained element can be present within
// the contained element.
func (opt *Options) OnChoice(ele xml.StartElement, protoTree []interface{}) (err error) {
	choice := Choice{}
	for _, attr := range ele.Attr {
		if attr.Name.Local == "maxOccurs" {
			var maxOccurs int
			if maxOccurs, err = strconv.Atoi(attr.Value); attr.Value != "unbounded" && err != nil {
				return
			}
			if attr.Value == "unbounded" || maxOccurs > 1 {
				choice.Plural, err = true, nil
			} else {
				choice.Plural, err = false, nil
			}
		}
	}
	// Handle a case of a parent choice having plurality that children should inherit
	if opt.Choice.Len() > 0 {
		choice.Plural = choice.Plural || opt.Choice.Peek().(*Choice).Plural
	}

	opt.Choice.Push(&choice)

	return
}

// EndChoice handles parsing event on the choice end elements.
func (opt *Options) EndChoice(ele xml.EndElement, protoTree []interface{}) (err error) {
	opt.Choice.Pop()

	return
}