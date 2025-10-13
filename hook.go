package xgen

import "encoding/xml"

// Hook is used for customizing the code generation process,
// by intercepting StartElement, EndElement, CharData and CodeGeneration
type Hook interface {
	OnStartElement(opt *Options, ele xml.StartElement, protoTree []interface{}) (next bool, err error)
	OnEndElement(opt *Options, ele xml.EndElement, protoTree []interface{}) (next bool, err error)
	OnCharData(opt *Options, ele string, protoTree []interface{}) (next bool, err error)
	OnGenerate(gen *CodeGenerator, protoName string, v interface{}) (next bool, err error)
}
