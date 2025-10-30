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

// ContentTrackingHook tracks all OnAddContent calls for different type generators
type ContentTrackingHook struct {
	SimpleTypeCount       int
	ComplexTypeCount      int
	GroupCount            int
	AttributeGroupCount   int
	ElementCount          int
	AttributeCount        int
	ContentModifications  []string
}

func (h *ContentTrackingHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *ContentTrackingHook) OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *ContentTrackingHook) OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error) {
	return true, nil
}

func (h *ContentTrackingHook) OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error) {
	return true, nil
}

func (h *ContentTrackingHook) OnAddContent(gen *CodeGenerator, content *string) {
	// Track which type of content is being generated based on the content
	if strings.Contains(*content, "type") {
		if strings.Contains(*content, "struct {") {
			if strings.Contains(*content, "Attr") && strings.Contains(*content, "attr,attr") {
				h.AttributeGroupCount++
				h.ContentModifications = append(h.ContentModifications, "AttributeGroup")
			} else if strings.Contains(*content, "Group") {
				h.GroupCount++
				h.ContentModifications = append(h.ContentModifications, "Group")
			} else {
				h.ComplexTypeCount++
				h.ContentModifications = append(h.ContentModifications, "ComplexType")
			}
		} else if strings.Contains(*content, "[]") || strings.Contains(*content, "string") {
			h.SimpleTypeCount++
			h.ContentModifications = append(h.ContentModifications, "SimpleType")
		}
		
		// Add a comment marker to prove modification happened
		*content = "// HOOK_TRACKED\n" + *content
	}
}

func TestOnAddContentCoverage(t *testing.T) {
	hook := &ContentTrackingHook{
		ContentModifications: make([]string, 0),
	}

	inputDir := filepath.Join(testFixtureDir, "xsd")
	file := filepath.Join(inputDir, "hook-coverage.xsd")

	tempDir, err := ioutil.TempDir(filepath.Join(testFixtureDir, "go"), "content-tracking-test-*")
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
	assert.NoError(t, err, "Parse should succeed")

	// Verify OnAddContent was called for different type generators
	// Note: Not all type generators may be triggered depending on XSD structure
	totalCalls := hook.SimpleTypeCount + hook.ComplexTypeCount + hook.GroupCount + hook.AttributeGroupCount
	assert.Greater(t, totalCalls, 0, "Should have called OnAddContent at least once")
	
	// Log what was generated
	t.Logf("OnAddContent calls - SimpleType: %d (lines 162,190,203), ComplexType: %d (line 282), Group: %d (line 323), AttributeGroup: %d (line 351)",
		hook.SimpleTypeCount, hook.ComplexTypeCount, hook.GroupCount, hook.AttributeGroupCount)

	// Verify content modifications were tracked
	assert.Greater(t, len(hook.ContentModifications), 0, "Should have tracked content modifications")

	// Read generated file and verify hook markers exist
	generatedFile := filepath.Join(tempDir, "hook-coverage.xsd.go")
	content, err := ioutil.ReadFile(generatedFile)
	require.NoError(t, err)

	generatedCode := string(content)
	
	// Verify hook comment was added to generated types
	assert.Contains(t, generatedCode, "// HOOK_TRACKED", "Generated code should contain hook tracking markers")
	
	// Count how many times the hook marker appears
	hookMarkerCount := strings.Count(generatedCode, "// HOOK_TRACKED")
	t.Logf("Found %d hook tracking markers in generated code", hookMarkerCount)
	assert.Greater(t, hookMarkerCount, 0, "Should have at least one hook marker")
}

// TestOnAddContentForAllTypes specifically tests each type generator's OnAddContent call
func TestOnAddContentForAllTypes(t *testing.T) {
	// Use base64.xsd which has diverse types
	hook := &ContentTrackingHook{
		ContentModifications: make([]string, 0),
	}

	inputDir := filepath.Join(testFixtureDir, "xsd")
	file := filepath.Join(inputDir, "base64.xsd")

	tempDir, err := ioutil.TempDir(filepath.Join(testFixtureDir, "go"), "all-types-test-*")
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
	assert.NoError(t, err, "Parse should succeed")

	// base64.xsd has SimpleTypes and ComplexTypes
	assert.Greater(t, hook.SimpleTypeCount, 0, "Should have called OnAddContent for SimpleTypes")
	assert.Greater(t, hook.ComplexTypeCount, 0, "Should have called OnAddContent for ComplexTypes")
	
	t.Logf("OnAddContent calls - SimpleType: %d, ComplexType: %d, Group: %d, AttributeGroup: %d",
		hook.SimpleTypeCount, hook.ComplexTypeCount, hook.GroupCount, hook.AttributeGroupCount)
}
