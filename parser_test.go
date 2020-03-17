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
	testDir     = "data"
	cSrcDir     = filepath.Join(testDir, "c")
	cCodeDir    = filepath.Join(cSrcDir, "output")
	goSrcDir    = filepath.Join(testDir, "go")
	goCodeDir   = filepath.Join(goSrcDir, "output")
	tsSrcDir    = filepath.Join(testDir, "ts")
	tsCodeDir   = filepath.Join(tsSrcDir, "output")
	javaSrcDir  = filepath.Join(testDir, "java")
	javaCodeDir = filepath.Join(javaSrcDir, "output")
	rsSrcDir    = filepath.Join(testDir, "rs")
	rsCodeDir   = filepath.Join(rsSrcDir, "output")
	xsdSrcDir   = filepath.Join(testDir, "xsd")
)

func TestParseGo(t *testing.T) {
	err := PrepareOutputDir(goCodeDir)
	assert.NoError(t, err)
	files, err := GetFileList(xsdSrcDir)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			OutputDir:           goCodeDir,
			Lang:                "Go",
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

func TestParseTypeScript(t *testing.T) {
	err := PrepareOutputDir(tsCodeDir)
	assert.NoError(t, err)
	files, err := GetFileList(xsdSrcDir)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			OutputDir:           tsCodeDir,
			Lang:                "TypeScript",
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err)
		if filepath.Ext(file) == ".xsd" {
			srcCode := filepath.Join(tsSrcDir, filepath.Base(file)+".ts")
			genCode := filepath.Join(tsCodeDir, filepath.Base(file)+".ts")

			srcFile, err := os.Stat(srcCode)
			assert.NoError(t, err)

			genFile, err := os.Stat(genCode)
			assert.NoError(t, err)

			assert.Equal(t, srcFile.Size(), genFile.Size(), fmt.Sprintf("error in generated code for %s", file))
		}
	}
}

func TestParseC(t *testing.T) {
	err := PrepareOutputDir(cCodeDir)
	assert.NoError(t, err)
	files, err := GetFileList(xsdSrcDir)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			OutputDir:           cCodeDir,
			Lang:                "C",
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err)
		if filepath.Ext(file) == ".xsd" {
			srcCode := filepath.Join(cSrcDir, filepath.Base(file)+".h")
			genCode := filepath.Join(cCodeDir, filepath.Base(file)+".h")

			srcFile, err := os.Stat(srcCode)
			assert.NoError(t, err)

			genFile, err := os.Stat(genCode)
			assert.NoError(t, err)

			assert.Equal(t, srcFile.Size(), genFile.Size(), fmt.Sprintf("error in generated code for %s", file))
		}
	}
}

func TestParseJava(t *testing.T) {
	err := PrepareOutputDir(javaCodeDir)
	assert.NoError(t, err)
	files, err := GetFileList(xsdSrcDir)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			OutputDir:           javaCodeDir,
			Lang:                "Java",
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err)
	}
}

func TestParseRust(t *testing.T) {
	err := PrepareOutputDir(rsCodeDir)
	assert.NoError(t, err)
	files, err := GetFileList(xsdSrcDir)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			OutputDir:           rsCodeDir,
			Lang:                "Rust",
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err)
	}
}
