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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testFixtureDir = "test"
	// externalFixtureDir is where one copy their own XSDs to run validation on them. For a set
	// of XSDs to run tests on, see https://github.com/xuri/xsd. Note that external tests leave the
	// generated output for inspection to support use-cases of manual review of generated code
	externalFixtureDir = "data"
)

func TestParseGo(t *testing.T) {
	testParseForSource(t, "Go", "go", "go", testFixtureDir, false, nil)
}

// TestParseGoExternal runs tests on any external XSDs within the externalFixtureDir
func TestParseGoExternal(t *testing.T) {
	testParseForSource(t, "Go", "go", "go", externalFixtureDir, true, nil)
}

// testParseForSource runs parsing tests for a given language. The sourceDirectory specifies the root of the
// input for the tests. The expected structure of the sourceDirectory is as follows:
//
//	source
//	├── xsd (with the input xsd files to run through the parser)
//	└── <langDirName> (with the expected generated code named <xsd-file>.<fileExt>
//
// The test cleans up files it generates unless leaveOutput is set to true. In which case, the generate file is left
// on disk for manual inspection under <sourceDirectory>/<langDirName>/output.
func testParseForSource(t *testing.T, lang string, fileExt string, langDirName string, sourceDirectory string, leaveOutput bool, hook Hook) {
	codeDir := filepath.Join(sourceDirectory, langDirName)

	outputDir := filepath.Join(codeDir, "output")
	if leaveOutput {
		err := PrepareOutputDir(outputDir)
		require.NoError(t, err)
	} else {
		tempDir, err := ioutil.TempDir(codeDir, "output-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		outputDir = tempDir
	}

	inputDir := filepath.Join(sourceDirectory, "xsd")
	files, err := GetFileList(inputDir)
	// Abort testing if the source directory doesn't include a xsd directory with inputs
	if os.IsNotExist(err) {
		return
	}

	require.NoError(t, err)
	for _, file := range files {
		if filepath.Ext(file) == ".xsd" {
			xsdName, err := filepath.Rel(inputDir, file)
			require.NoError(t, err)

			t.Run(xsdName, func(t *testing.T) {
				parser := NewParser(&Options{
					FilePath:            file,
					InputDir:            inputDir,
					OutputDir:           outputDir,
					Lang:                lang,
					IncludeMap:          make(map[string]bool),
					LocalNameNSMap:      make(map[string]string),
					NSSchemaLocationMap: make(map[string]string),
					ParseFileList:       make(map[string]bool),
					ParseFileMap:        make(map[string][]interface{}),
					ProtoTree:           make([]interface{}, 0),
					Hook:                hook,
				})
				err = parser.Parse()
				assert.NoError(t, err, file)
				generatedFileName := strings.TrimPrefix(file, inputDir) + "." + fileExt
				actualFilename := filepath.Join(outputDir, generatedFileName)

				actualGenerated, err := ioutil.ReadFile(actualFilename)
				assert.NoError(t, err)

				expectedFilename := filepath.Join(codeDir, generatedFileName)
				expectedGenerated, err := ioutil.ReadFile(expectedFilename)
				assert.NoError(t, err)

				assert.Equal(t, string(expectedGenerated), string(actualGenerated), fmt.Sprintf("error in generated code for %s", file))
			})
		}
	}
}

func TestParseTypeScript(t *testing.T) {
	testParseForSource(t, "TypeScript", "ts", "ts", testFixtureDir, false, nil)
}

func TestParseTypeScriptExternal(t *testing.T) {
	testParseForSource(t, "TypeScript", "ts", "ts", externalFixtureDir, true, nil)
}

func TestParseC(t *testing.T) {
	testParseForSource(t, "C", "h", "c", testFixtureDir, false, nil)
}

func TestParseCExternal(t *testing.T) {
	testParseForSource(t, "C", "h", "c", externalFixtureDir, true, nil)
}

func TestParseJava(t *testing.T) {
	testParseForSource(t, "Java", "java", "java", testFixtureDir, false, nil)
}

func TestParseJavaExternal(t *testing.T) {
	testParseForSource(t, "Java", "java", "java", externalFixtureDir, true, nil)
}

func TestParseRust(t *testing.T) {
	testParseForSource(t, "Rust", "rs", "rs", testFixtureDir, false, nil)
}

func TestParseRustExternal(t *testing.T) {
	testParseForSource(t, "Rust", "rs", "rs", externalFixtureDir, true, nil)
}

type Appinfo struct {
	Doc    string
	Parent string
}

type AppinfoHook struct {
	Override          bool
	OnStartElementRan bool
	OnEndElementRan   bool
	OnCharDataRan     bool
	OnGenerateRan     bool

	Appinfo *Stack
}

func (h *AppinfoHook) ShouldOverride() bool {
	return h.Override
}

func (h *AppinfoHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error) {
	if ele.Name.Local != "appinfo" {
		return true, nil
	}

	h.OnStartElementRan = true

	a := &Appinfo{}

	a.Parent = opt.CurrentEle

	if opt.InElement != "" && opt.Element.Peek() != nil {
		a.Parent = opt.Element.Peek().(*Element).Name
	}

	switch opt.CurrentEle {
	case "simpleType":
		if opt.SimpleType.Peek() != nil {
			a.Parent = opt.SimpleType.Peek().(*SimpleType).Name
		}
	case "complexType":
		if opt.ComplexType.Peek() != nil {
			a.Parent = opt.ComplexType.Peek().(*ComplexType).Name
		}
	case "element":
		if opt.Element.Peek() != nil {
			a.Parent = opt.Element.Peek().(*Element).Name
		}
	}

	h.Appinfo.Push(a)

	return true, nil
}

func (h *AppinfoHook) OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error) {
	if ele.Name.Local == "appinfo" {
		return true, nil
	}

	h.OnEndElementRan = true
	opt.ProtoTree = append(opt.ProtoTree, h.Appinfo.Pop())

	return true, nil
}

func (h *AppinfoHook) OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error) {
	if h.Appinfo.Peek() != nil {
		h.OnCharDataRan = true
		h.Appinfo.Peek().(*Appinfo).Doc = ele
	}
	return true, nil
}

func (h *AppinfoHook) OnGenerate(gen *CodeGenerator, protoName string, ele interface{}) (next bool, err error) {
	h.OnGenerateRan = false
	switch v := ele.(type) {
	case *ComplexType:
		if _, ok := gen.StructAST[v.Name]; !ok {
			// for this fixture, at least one attribute must exist, and must have a name
			for _, attribute := range v.Attributes {
				h.OnGenerateRan = h.OnGenerateRan || attribute.Name != ""
			}
		}
	}
	return true, nil
}

func (h *AppinfoHook) OnAddContent(gen *CodeGenerator, content *string) {
	// no-op
}

func TestParseGoWithAppinfoHook(t *testing.T) {
	appinfoHook := &AppinfoHook{}
	appinfoHook.Appinfo = NewStack()
	testParseForSource(t, "Go", "go", "go", testFixtureDir, false, appinfoHook)
	assert.True(t, appinfoHook.OnStartElementRan)
	assert.True(t, appinfoHook.OnEndElementRan)
	assert.True(t, appinfoHook.OnCharDataRan)
	assert.True(t, appinfoHook.OnGenerateRan)
}

// ComprehensiveHook tests skipping elements, filtering generation, and content modification
type ComprehensiveHook struct {
	SkippedAnnotations int
	SkippedTypes       []string
	ModifiedContent    bool
}

func (h *ComprehensiveHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (bool, error) {
	// Skip all <annotation> elements to test filtering behavior
	if ele.Name.Local == "annotation" {
		h.SkippedAnnotations++
		return false, nil // Skip processing this element
	}
	return true, nil
}

func (h *ComprehensiveHook) OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *ComprehensiveHook) OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *ComprehensiveHook) OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error) {
	// Skip generating code for SimpleType named "myType1" to test generation filtering
	if protoName == "SimpleType" {
		if st, ok := v.(*SimpleType); ok && st.Name == "myType1" {
			h.SkippedTypes = append(h.SkippedTypes, st.Name)
			return false, nil // Skip generating this type
		}
	}
	return true, nil
}

func (h *ComprehensiveHook) OnAddContent(gen *CodeGenerator, content *string) {
	// Modify generated content to add a custom marker comment
	if strings.Contains(*content, "type MyType2") {
		*content = strings.Replace(*content, "type MyType2", "// HOOK_MODIFIED\ntype MyType2", 1)
		h.ModifiedContent = true
	}
}

func TestHookSkipAndModify(t *testing.T) {
	hook := &ComprehensiveHook{
		SkippedTypes: make([]string, 0),
	}

	// Create temp directory for output
	tempDir, err := ioutil.TempDir(filepath.Join(testFixtureDir, "go"), "hook-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Run parser with hook
	inputDir := filepath.Join(testFixtureDir, "xsd")
	files, err := GetFileList(inputDir)
	require.NoError(t, err)

	for _, file := range files {
		if filepath.Ext(file) == ".xsd" {
			parser := NewParser(&Options{
				FilePath:            file,
				InputDir:            inputDir,
				OutputDir:           tempDir,
				Lang:                "Go",
				IncludeMap:          make(map[string]bool),
				LocalNameNSMap:      make(map[string]string),
				NSSchemaLocationMap: make(map[string]string),
				ParseFileList:       make(map[string]bool),
				ParseFileMap:        make(map[string][]interface{}),
				ProtoTree:           make([]interface{}, 0),
				Hook:                hook,
			})
			err = parser.Parse()
			assert.NoError(t, err)
		}
	}

	// Verify skipping worked - annotations in base64.xsd should have been skipped
	assert.Greater(t, hook.SkippedAnnotations, 0, "Hook should have skipped annotation elements")

	// Verify type filtering worked
	assert.Contains(t, hook.SkippedTypes, "myType1", "Hook should have skipped myType1 generation")

	// Read generated file and verify modifications
	generatedFile := filepath.Join(tempDir, "base64.xsd.go")
	content, err := ioutil.ReadFile(generatedFile)
	require.NoError(t, err)

	generatedCode := string(content)

	// Verify hook comment was added
	if hook.ModifiedContent {
		assert.Contains(t, generatedCode, "// HOOK_MODIFIED", "Generated code should contain hook modifications")
	}

	// Verify skipped type was not generated
	// MyType1 should not have its own type declaration (it's used as a field type but shouldn't have "type MyType1 ")
	assert.NotContains(t, generatedCode, "type MyType1 ", "Skipped type should not have its own type declaration")

	// Verify it's still referenced in TopLevel (the field name, not the type declaration)
	assert.Contains(t, generatedCode, "MyType1         []string", "Field referencing the type should still exist")

	// Verify other types were still generated
	assert.Contains(t, generatedCode, "type MyType2", "Non-skipped types should be generated")
}

// ErrorTestHook tests error propagation from hooks
type ErrorTestHook struct{}

func (h *ErrorTestHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error) {
	// Return error when encountering specific element
	if ele.Name.Local == "complexType" {
		for _, attr := range ele.Attr {
			if attr.Name.Local == "name" && attr.Value == "myType3" {
				return false, fmt.Errorf("intentional error for testing: forbidden type myType3")
			}
		}
	}
	return true, nil
}

func (h *ErrorTestHook) OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *ErrorTestHook) OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *ErrorTestHook) OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error) {
	return true, nil
}

func (h *ErrorTestHook) OnAddContent(gen *CodeGenerator, content *string) {
	// no-op
}

func TestHookErrorHandling(t *testing.T) {
	hook := &ErrorTestHook{}

	inputDir := filepath.Join(testFixtureDir, "xsd")
	file := filepath.Join(inputDir, "base64.xsd")

	tempDir, err := ioutil.TempDir(filepath.Join(testFixtureDir, "go"), "error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	parser := NewParser(&Options{
		FilePath:            file,
		InputDir:            inputDir,
		OutputDir:           tempDir,
		Lang:                "Go",
		IncludeMap:          make(map[string]bool),
		LocalNameNSMap:      make(map[string]string),
		NSSchemaLocationMap: make(map[string]string),
		ParseFileList:       make(map[string]bool),
		ParseFileMap:        make(map[string][]interface{}),
		ProtoTree:           make([]interface{}, 0),
		Hook:                hook,
	})

	err = parser.Parse()

	// Verify that the error from the hook was propagated
	assert.Error(t, err, "Hook error should be propagated")
	assert.Contains(t, err.Error(), "intentional error for testing", "Error should contain hook's error message")
	assert.Contains(t, err.Error(), "forbidden type myType3", "Error should identify the specific type")
}

// CharDataErrorHook tests error handling in OnCharData
type CharDataErrorHook struct {
	EncounteredCharData bool
}

func (h *CharDataErrorHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *CharDataErrorHook) OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *CharDataErrorHook) OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error) {
	// Return error on non-empty character data
	trimmed := strings.TrimSpace(ele)
	if trimmed != "" {
		h.EncounteredCharData = true
		return false, fmt.Errorf("intentional OnCharData error: got data '%s'", trimmed)
	}
	return true, nil
}

func (h *CharDataErrorHook) OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error) {
	return true, nil
}

func (h *CharDataErrorHook) OnAddContent(gen *CodeGenerator, content *string) {
	// no-op
}

func TestHookOnCharDataError(t *testing.T) {
	hook := &CharDataErrorHook{}

	inputDir := filepath.Join(testFixtureDir, "xsd")
	file := filepath.Join(inputDir, "base64.xsd")

	tempDir, err := ioutil.TempDir(filepath.Join(testFixtureDir, "go"), "chardata-error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	parser := NewParser(&Options{
		FilePath:            file,
		InputDir:            inputDir,
		OutputDir:           tempDir,
		Lang:                "Go",
		IncludeMap:          make(map[string]bool),
		LocalNameNSMap:      make(map[string]string),
		NSSchemaLocationMap: make(map[string]string),
		ParseFileList:       make(map[string]bool),
		ParseFileMap:        make(map[string][]interface{}),
		ProtoTree:           make([]interface{}, 0),
		Hook:                hook,
	})

	err = parser.Parse()

	// Verify that the error from OnCharData was propagated
	assert.Error(t, err, "OnCharData error should be propagated")
	assert.Contains(t, err.Error(), "intentional OnCharData error", "Error should contain OnCharData's error message")
	assert.True(t, hook.EncounteredCharData, "Hook should have encountered character data")
}

// CharDataSkipHook tests skipping behavior in OnCharData
type CharDataSkipHook struct {
	SkippedCharData int
	ProcessedCount  int
}

func (h *CharDataSkipHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *CharDataSkipHook) OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *CharDataSkipHook) OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error) {
	h.ProcessedCount++
	// Skip processing for character data containing "appinfo"
	if strings.Contains(ele, "appinfo") {
		h.SkippedCharData++
		return false, nil // Skip further processing
	}
	return true, nil
}

func (h *CharDataSkipHook) OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error) {
	return true, nil
}

func (h *CharDataSkipHook) OnAddContent(gen *CodeGenerator, content *string) {
	// no-op
}

func TestHookOnCharDataSkip(t *testing.T) {
	hook := &CharDataSkipHook{}

	inputDir := filepath.Join(testFixtureDir, "xsd")
	file := filepath.Join(inputDir, "base64.xsd")

	tempDir, err := ioutil.TempDir(filepath.Join(testFixtureDir, "go"), "chardata-skip-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	parser := NewParser(&Options{
		FilePath:            file,
		InputDir:            inputDir,
		OutputDir:           tempDir,
		Lang:                "Go",
		IncludeMap:          make(map[string]bool),
		LocalNameNSMap:      make(map[string]string),
		NSSchemaLocationMap: make(map[string]string),
		ParseFileList:       make(map[string]bool),
		ParseFileMap:        make(map[string][]interface{}),
		ProtoTree:           make([]interface{}, 0),
		Hook:                hook,
	})

	err = parser.Parse()

	// Parsing should succeed even though we skipped some character data
	assert.NoError(t, err, "Parse should succeed when skipping character data")

	// Verify that the hook processed character data and skipped some
	assert.Greater(t, hook.ProcessedCount, 0, "Hook should have processed character data")
	assert.Greater(t, hook.SkippedCharData, 0, "Hook should have skipped some character data")
}

// OnAddContentTestHook tracks OnAddContent calls
type OnAddContentTestHook struct {
	OnAddContentCallCount int
}

func (h *OnAddContentTestHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *OnAddContentTestHook) OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *OnAddContentTestHook) OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *OnAddContentTestHook) OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error) {
	return true, nil
}

func (h *OnAddContentTestHook) OnAddContent(gen *CodeGenerator, content *string) {
	h.OnAddContentCallCount++
}

// TestOnAddContentHookNotNil tests all locations where: if gen.Hook != nil { gen.Hook.OnAddContent(gen, &output) }
// This covers the TRUE branch (gen.Hook != nil)
func TestOnAddContentHookNotNil(t *testing.T) {
	fieldNameCount = make(map[string]int)

	hook := &OnAddContentTestHook{}
	gen := &CodeGenerator{
		Lang:      "Go",
		StructAST: make(map[string]string),
		Hook:      hook, // NOT nil
	}

	tests := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "GoSimpleType_List",
			testFunc: func() {
				gen.GoSimpleType(&SimpleType{Name: "ListType", Base: "string", List: true})
			},
		},
		{
			name: "GoSimpleType_Union",
			testFunc: func() {
				gen.GoSimpleType(&SimpleType{
					Name:        "UnionType",
					Union:       true,
					MemberTypes: map[string]string{"string": "string", "int": "int"},
				})
			},
		},
		{
			name: "GoSimpleType_Base",
			testFunc: func() {
				gen.GoSimpleType(&SimpleType{Name: "BaseType", Base: "string"})
			},
		},
		{
			name: "GoComplexType",
			testFunc: func() {
				gen.GoComplexType(&ComplexType{
					Name:     "ComplexType",
					Elements: []Element{{Name: "Field1", Type: "string"}},
				})
			},
		},
		{
			name: "GoGroup",
			testFunc: func() {
				gen.GoGroup(&Group{
					Name:     "GroupType",
					Elements: []Element{{Name: "Element1", Type: "string"}},
				})
			},
		},
		{
			name: "GoAttributeGroup",
			testFunc: func() {
				gen.GoAttributeGroup(&AttributeGroup{
					Name:       "AttrGroupType",
					Attributes: []Attribute{{Name: "Attr1", Type: "string"}},
				})
			},
		},
		{
			name: "GoElement",
			testFunc: func() {
				gen.GoElement(&Element{Name: "ElementType", Type: "string"})
			},
		},
		{
			name: "GoAttribute",
			testFunc: func() {
				gen.GoAttribute(&Attribute{Name: "AttributeType", Type: "string"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := hook.OnAddContentCallCount
			tt.testFunc()
			assert.Greater(t, hook.OnAddContentCallCount, initialCount,
				"OnAddContent should be called when Hook is not nil for %s", tt.name)
		})
	}
}

// TestOnAddContentHookIsNil tests all locations where: if gen.Hook != nil { gen.Hook.OnAddContent(gen, &output) }
// This covers the FALSE branch (gen.Hook == nil)
func TestOnAddContentHookIsNil(t *testing.T) {
	fieldNameCount = make(map[string]int)

	gen := &CodeGenerator{
		Lang:      "Go",
		StructAST: make(map[string]string),
		Hook:      nil, // IS nil
	}

	tests := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "GoSimpleType_List",
			testFunc: func() {
				gen.GoSimpleType(&SimpleType{Name: "ListType", Base: "string", List: true})
			},
		},
		{
			name: "GoSimpleType_Union",
			testFunc: func() {
				gen.GoSimpleType(&SimpleType{
					Name:        "UnionType",
					Union:       true,
					MemberTypes: map[string]string{"string": "string", "int": "int"},
				})
			},
		},
		{
			name: "GoSimpleType_Base",
			testFunc: func() {
				gen.GoSimpleType(&SimpleType{Name: "BaseType", Base: "string"})
			},
		},
		{
			name: "GoComplexType",
			testFunc: func() {
				gen.GoComplexType(&ComplexType{
					Name:     "ComplexType",
					Elements: []Element{{Name: "Field1", Type: "string"}},
				})
			},
		},
		{
			name: "GoGroup",
			testFunc: func() {
				gen.GoGroup(&Group{
					Name:     "GroupType",
					Elements: []Element{{Name: "Element1", Type: "string"}},
				})
			},
		},
		{
			name: "GoAttributeGroup",
			testFunc: func() {
				gen.GoAttributeGroup(&AttributeGroup{
					Name:       "AttrGroupType",
					Attributes: []Attribute{{Name: "Attr1", Type: "string"}},
				})
			},
		},
		{
			name: "GoElement",
			testFunc: func() {
				gen.GoElement(&Element{Name: "ElementType", Type: "string"})
			},
		},
		{
			name: "GoAttribute",
			testFunc: func() {
				gen.GoAttribute(&Attribute{Name: "AttributeType", Type: "string"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.testFunc,
				"Should not panic when Hook is nil for %s", tt.name)
		})
	}
}
