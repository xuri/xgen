<p align="center"><img width="450" src="./xgen.svg" alt="xgen logo"></p>

<br>

<p align="center">
    <a href="https://github.com/xuri/xgen/actions/workflows/go.yml"><img src="https://github.com/xuri/xgen/actions/workflows/go.yml/badge.svg" alt="Build Status"></a>
    <a href="https://codecov.io/gh/xuri/xgen"><img src="https://codecov.io/gh/xuri/xgen/branch/master/graph/badge.svg" alt="Code Coverage"></a>
    <a href="https://goreportcard.com/report/github.com/xuri/xgen"><img src="https://goreportcard.com/badge/github.com/xuri/xgen" alt="Go Report Card"></a>
    <a href="https://pkg.go.dev/github.com/xuri/xgen"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="go.dev"></a>
    <a href="https://opensource.org/licenses/BSD-3-Clause"><img src="https://img.shields.io/badge/license-bsd-orange.svg" alt="Licenses"></a>
    <a href="https://www.paypal.me/xuri"><img src="https://img.shields.io/badge/Donate-PayPal-green.svg" alt="Donate"></a>
</p>

# xgen

## Introduction

xgen is a library written in pure Go providing a set of functions that allow you to parse XSD (XML schema definition) files. This library needs Go version 1.23.0 or later. The full API docs can be seen using go's built-in documentation tool, or online at [go.dev](https://pkg.go.dev/github.com/xuri/xgen).

`xgen` commands automatically compiles XML schema files into the multi-language type or class declarations code.

Install the command line tool first, installing executables with `go install`

```sh
go install github.com/xuri/xgen/cmd/xgen@latest
```

The command below will walk on the `xsd` path and generate Go language struct code under the `output` directory.

```text
$ xgen -i /path/to/your/xsd -o /path/to/your/output -l Go
```

Usage:

```text
$ xgen [<flag> ...] <XSD file or directory> ...
   -i <path> Input file path or directory for the XML schema definition
   -o <path> Output file path or directory for the generated code
   -p        Specify the package name
   -l        Specify the language of generated code (Go/C/Java/Rust/TypeScript)
   -h        Output this help and exit
   -v        Output version and exit
```

## Programmatic Usage

You can use xgen as a library in your Go code for more control over the parsing and code generation process.

### Basic Usage

```go
import "github.com/xuri/xgen"

parser := xgen.NewParser(&xgen.Options{
    FilePath:            "schema.xsd",
    OutputDir:           "output",
    Lang:                "Go",
    Package:             "mypackage",
    IncludeMap:          make(map[string]bool),
    LocalNameNSMap:      make(map[string]string),
    NSSchemaLocationMap: make(map[string]string),
    ParseFileList:       make(map[string]bool),
    ParseFileMap:        make(map[string][]interface{}),
    ProtoTree:           make([]interface{}, 0),
})
err := parser.Parse()
```

### Customization with Hooks

The `Hook` interface allows you to customize the parsing and code generation process by intercepting events at various stages:

```go
type CustomHook struct{}

func (h *CustomHook) OnStartElement(opt *xgen.Options, ele xml.StartElement, protoTree []interface{}) (bool, error) {
    // Intercept XML elements during parsing
    return true, nil
}

func (h *CustomHook) OnGenerate(gen *xgen.CodeGenerator, protoName string, v interface{}) (bool, error) {
    // Intercept code generation for each type
    return true, nil
}

func (h *CustomHook) OnAddContent(gen *xgen.CodeGenerator, content *string) {
    // Modify generated code before writing to file
}

// ... implement other Hook methods ...

parser := xgen.NewParser(&xgen.Options{
    // ... other options ...
    Hook: &CustomHook{},
})
```

Use cases for hooks include:
- Parsing custom XSD extensions or vendor-specific annotations
- Customizing type mappings between XSD and target language types
- Injecting additional methods or documentation into generated code
- Filtering elements during parsing or code generation

See the `Hook` interface documentation and `TestParseGoWithAppinfoHook` in `parser_test.go` for complete examples.

## XSD (XML Schema Definition)

XSD, a recommendation of the World Wide Web Consortium ([W3C](https://www.w3.org)), specifies how to formally describe the elements in an Extensible Markup Language ([XML](https://www.w3.org/TR/xml/)) document. It can be used by programmers to verify each piece of item content in a document. They can check if it adheres to the description of the element it is placed in.

XSD can be used to express a set of rules to which an XML document must conform in order to be considered "valid" according to that schema. However, unlike most other schema languages, XSD was also designed with the intent that determination of a document's validity would produce a collection of information adhering to specific data types. Such a post-validation infoset can be useful in the development of XML document processing software.

## Contributing

Contributions are welcome! Open a pull request to fix a bug, or open an issue to discuss a new feature or change. XSD is compliant with [XML Schema Part 1: Structures Second Edition](https://www.w3.org/TR/xmlschema-1/).

## Licenses

This program is under the terms of the BSD 3-Clause License. See [https://opensource.org/licenses/BSD-3-Clause](https://opensource.org/licenses/BSD-3-Clause).

Logo is designed by [xuri](https://xuri.me). Licensed under the [Creative Commons 3.0 Attributions license](http://creativecommons.org/licenses/by/3.0/).
