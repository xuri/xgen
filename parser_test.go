// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	assert.NoError(t, err)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			InputDir:            xsdSrcDir,
			OutputDir:           goCodeDir,
			Lang:                "Go",
			IncludeMap:          make(map[string]bool),
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err, file)
		if filepath.Ext(file) == ".xsd" {
			srcCode := filepath.Join(goSrcDir, strings.TrimPrefix(file, xsdSrcDir)+".go")
			genCode := filepath.Join(goCodeDir, strings.TrimPrefix(file, xsdSrcDir)+".go")

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
	assert.NoError(t, err)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			InputDir:            xsdSrcDir,
			OutputDir:           tsCodeDir,
			Lang:                "TypeScript",
			IncludeMap:          make(map[string]bool),
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err)
		if filepath.Ext(file) == ".xsd" {
			srcCode := filepath.Join(tsSrcDir, strings.TrimPrefix(file, xsdSrcDir)+".ts")
			genCode := filepath.Join(tsCodeDir, strings.TrimPrefix(file, xsdSrcDir)+".ts")

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
	assert.NoError(t, err)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			InputDir:            xsdSrcDir,
			OutputDir:           cCodeDir,
			Lang:                "C",
			IncludeMap:          make(map[string]bool),
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
			ProtoTree:           make([]interface{}, 0),
		})
		err = parser.Parse()
		assert.NoError(t, err)
		if filepath.Ext(file) == ".xsd" {
			srcCode := filepath.Join(cSrcDir, strings.TrimPrefix(file, xsdSrcDir)+".h")
			genCode := filepath.Join(cCodeDir, strings.TrimPrefix(file, xsdSrcDir)+".h")

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
	assert.NoError(t, err)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			InputDir:            xsdSrcDir,
			OutputDir:           javaCodeDir,
			Lang:                "Java",
			IncludeMap:          make(map[string]bool),
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
	assert.NoError(t, err)
	for _, file := range files {
		parser := NewParser(&Options{
			FilePath:            file,
			InputDir:            xsdSrcDir,
			OutputDir:           rsCodeDir,
			Lang:                "Rust",
			IncludeMap:          make(map[string]bool),
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
