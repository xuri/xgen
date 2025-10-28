package xgen

import "encoding/xml"

// Hook provides a mechanism for customizing the XSD parsing and code generation process
// by intercepting events at various stages. Implementations can filter, modify, or extend
// the default behavior without modifying xgen's core logic.
//
// Use cases include:
//   - Parsing custom XSD extensions or vendor-specific annotations
//   - Filtering elements to skip during parsing or code generation
//   - Customizing type mappings between XSD and target language types
//   - Injecting additional code, methods, or documentation into generated output
//   - Logging and debugging the parsing/generation flow
//
// Example:
//
//	type CustomHook struct{}
//
//	func (h *CustomHook) OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error) {
//	    if ele.Name.Local == "customElement" {
//	        // Handle custom XSD element
//	        return false, nil // Skip default processing
//	    }
//	    return true, nil // Continue with default processing
//	}
//
//	// ... implement other Hook methods ...
//
//	parser := NewParser(&Options{
//	    FilePath: "schema.xsd",
//	    OutputDir: "output",
//	    Lang: "Go",
//	    Hook: &CustomHook{},
//	})
//
// For a complete working example, see TestParseGoWithAppinfoHook in parser_test.go.
type Hook interface {
	// OnStartElement is called when an XML start element is encountered during parsing.
	// Return next=false to skip xgen's default processing of this element.
	// Return an error to halt parsing.
	OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error)

	// OnEndElement is called when an XML end element is encountered during parsing.
	// Return next=false to skip xgen's default processing of this element.
	// Return an error to halt parsing.
	OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error)

	// OnCharData is called when XML character data (text content) is encountered during parsing.
	// Return next=false to skip xgen's default processing of this content.
	// Return an error to halt parsing.
	OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error)

	// OnGenerate is called before generating code for each type (SimpleType, ComplexType, etc.).
	// The protoName identifies the type being generated (e.g., "SimpleType", "ComplexType").
	// Return next=false to skip code generation for this type.
	// Return an error to halt code generation.
	OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error)

	// OnAddContent is called after each code block is generated, allowing modification
	// of the generated code. The content parameter is a pointer to the generated code string
	// and can be modified directly.
	OnAddContent(gen *CodeGenerator, content *string)
}
