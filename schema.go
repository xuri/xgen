// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "encoding/xml"

func (opt *Options) prepareLocalNameNSMap(element xml.StartElement) {
	for _, ele := range element.Attr {
		if ele.Name.Space == "xmlns" {
			opt.LocalNameNSMap[ele.Name.Local] = ele.Value
		}
	}
	return
}

func (opt *Options) prepareNSSchemaLocationMap(element xml.StartElement) {
	var currentNS string
	for _, ele := range element.Attr {
		if ele.Name.Local == "namespace" {
			currentNS = ele.Value
		}
		if ele.Name.Local == "schemaLocation" {
			if _, ok := opt.NSSchemaLocationMap[currentNS]; ok {
				continue
			}
			if isValidURL(ele.Value) {
				continue
				// TODO: fetch remote schema
				// var err error
				// if opt.RemoteSchema[ele.Value], err = fetchSchema(ele.Value); err != nil {
				// 	continue
				// }
			}
			opt.NSSchemaLocationMap[currentNS] = ele.Value
		}
	}
	return
}

func (opt *Options) parseNS(str string) (ns string) {
	return opt.LocalNameNSMap[getNSPrefix(str)]
}
