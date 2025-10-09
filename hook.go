package xgen

import "encoding/xml"

// Hook is used for customizing the code generation process,
// by intercepting StartElement, EndElement, CharData and CodeGeneration
type Hook interface {
	ShouldOverride() bool
	OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (err error)
	OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (err error)
	OnCharData(opt *Options, ele string, protoTree []interface{}) (err error)
	OnGenerate(gen *CodeGenerator, v interface{})
}
