// Copyright 2020 - 2024 The xgen Authors. All rights reserved. Use of this
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

// OnSequence evaluates wether the sequence element contains a maxOccurs attribute
// (that in turn mandates plural inner elements) and saves that info on a stack
func (opt *Options) OnSequence(ele xml.StartElement, protoTree []interface{}) (err error) {
	for _, attr := range ele.Attr {
		if attr.Name.Local == "maxOccurs" {
			if attr.Value == "unbounded" {
				opt.InPluralSequence = append(opt.InPluralSequence, true)
				return nil
			}

			var maxOccurs int
			if maxOccurs, err = strconv.Atoi(attr.Value); err == nil && maxOccurs > 1 {
				opt.InPluralSequence = append(opt.InPluralSequence, true)
				return nil
			}
		}
	}
	opt.InPluralSequence = append(opt.InPluralSequence, false)
	return nil
}

// EndSequence removes an item from the stack mentioned above
func (opt *Options) EndSequence(ele xml.EndElement, protoTree []interface{}) (err error) {
	opt.InPluralSequence = opt.InPluralSequence[:len(opt.InPluralSequence)-1]
	return
}
