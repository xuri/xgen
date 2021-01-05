// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

// OnImport handles parsing event on the import start elements. The list
// element defines a simple type element as a list of values of a specified
// data type.
func (opt *Options) OnImport(ele xml.StartElement, protoTree []interface{}) (err error) {
	opt.prepareNSSchemaLocationMap(ele)
	return
}
