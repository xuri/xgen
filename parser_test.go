// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	goSrcDir  = filepath.Join("test", "go")
	goCodeDir = filepath.Join(goSrcDir, "output")
	xsdSrcDir = filepath.Join("test", "xsd")
)

func TestParse(t *testing.T) {
	err := PrepareOutputDir(goCodeDir)
	assert.NoError(t, err)
	files, err := GetFileList(xsdSrcDir)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			OutputDir:           goCodeDir,
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err)
		if filepath.Ext(file) == ".xsd" {
			srcCode := filepath.Join(goSrcDir, filepath.Base(file)+".go")
			genCode := filepath.Join(goCodeDir, filepath.Base(file)+".go")

			srcFile, err := os.Stat(srcCode)
			assert.NoError(t, err)

			genFile, err := os.Stat(genCode)
			assert.NoError(t, err)

			assert.Equal(t, srcFile.Size(), genFile.Size(), fmt.Sprintf("error in generated code for %s", file))
		}

	}

}
