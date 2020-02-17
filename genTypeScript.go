// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "os"

// GenTypeScript generate TypeScript programming language source code for XML
// schema definition files.
func (gen *CodeGenerator) GenTypeScript() error {
	f, err := os.Create(gen.File + ".ts")
	if err != nil {
		return err
	}
	defer f.Close()
	return err
}
