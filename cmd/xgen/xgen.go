// Copyright 2020 The xgen Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE
// file.
//
// xgen is a tool to automatically compiles XML schema files into the
// multi-language type or class declarations code.
//
// Usage:
//
//    $ xgen [<flag> ...] <XSD file or directory> ...
//        -i <path> Input file path or directory for the XML schema definition
//        -o <path> Output file path or directory for the generated code
//        -p        Specify the package name
//        -l        Specify the language of generated code (Go/C/Java/Rust/TypeScript)
//        -h        Output this help and exit
//        -v        Output version and exit
//
// If the path specified by the -i flag is a directory, all files in the
// directory will be processed as XML schema definition.
//
// The default package name and output directory are "schema" and "xgen_out".
//
// Currently support language is Go.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xuri/xgen"
)

// Config holds user-defined overrides and filters that are used when
// generating source code from an XSD document.
type Config struct {
	I       string
	O       string
	Pkg     string
	Lang    string
	Version string
}

// Cfg are the default config for xgen. The default package name and output
// directory are "schema" and "xgen_out".
var Cfg = Config{
	Pkg:     "schema",
	Version: "0.1.0",
}

// SupportLang defines supported language types.
var SupportLang = map[string]bool{
	"Go":         true,
	"C":          false,
	"Java":       false,
	"Rust":       false,
	"TypeScript": false,
}

// parseFlags parse flags of program.
func parseFlags() *Config {
	iPtr := flag.String("i", "", "Input file path or directory for the XML schema definition")
	oPtr := flag.String("o", "xgen_out", "Output file path or directory for the generated code")
	pkgPtr := flag.String("p", "", "Specify the package name")
	langPtr := flag.String("l", "", "Specify the language of generated code")
	verPtr := flag.Bool("v", false, "Show version and exit")
	helpPtr := flag.Bool("h", false, "Show this help and exit")
	flag.Parse()
	if *helpPtr {
		fmt.Printf("xgen version: %s\r\nCopyright (c) 2020 Ri Xu https://xuri.me All rights reserved.\r\n\r\nUsage:\r\n$ xgen [<flag> ...] <XSD file or directory> ...\n  -i <path>\tInput file path or directory for the XML schema definition\r\n  -o <path>\tOutput file path or directory for the generated code\r\n  -p     \tSpecify the package name\r\n  -l      \tSpecify the language of generated code (Go/C/Java/Rust/TypeScript)\r\n  -h     \tOutput this help and exit\r\n  -v     \tOutput version and exit\r\n", Cfg.Version)
		os.Exit(0)
	}
	if *verPtr {
		fmt.Printf("xgen version: %s\r\n", Cfg.Version)
		os.Exit(0)
	}
	if *iPtr == "" {
		fmt.Println("must specify input file path or directory for the XML schema definition")
		os.Exit(1)
	}
	Cfg.I = *iPtr
	if *langPtr == "" {
		fmt.Println("must specify the language of generated code (Go/C/Java/Rust/TypeScript)")
		os.Exit(1)
	}
	Cfg.Lang = *langPtr
	if *oPtr != "" {
		Cfg.O = *oPtr
	}
	if ok := SupportLang[Cfg.Lang]; !ok {
		fmt.Println("not support language", Cfg.Lang)
		os.Exit(1)
	}
	if *pkgPtr != "" {
		Cfg.Pkg = *pkgPtr
	}
	return &Cfg
}

func main() {
	cfg := parseFlags()
	if err := xgen.PrepareOutputDir(cfg.O); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	files, err := xgen.GetFileList(cfg.I)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, file := range files {
		parser := xgen.NewParser(&xgen.Options{
			FilePath:            file,
			OutputDir:           cfg.O,
			LocalNameNSMap:      make(map[string]string),
			NSSchemaLocationMap: make(map[string]string),
			ParseFileList:       make(map[string]bool),
			ParseFileMap:        make(map[string][]interface{}),
		})
		if _, err = parser.Parse(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	fmt.Println("done")
}
