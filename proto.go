// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "regexp"

// SimpleType definitions provide for constraining character information item
// [children] of element and attribute information items.
// https://www.w3.org/TR/xmlschema-1/#Simple_Type_Definitions
type SimpleType struct {
	Doc         string
	Name        string
	Base        string
	Anonymous   bool
	List        bool
	Union       bool
	MemberTypes map[string]string
	Restriction Restriction
}

// Element declarations provide for: Local validation of element information
// item values using a type definition; Specifying default or fixed values for
// an element information items; Establishing uniquenesses and reference
// constraint relationships among the values of related elements and
// attributes; Controlling the substitutability of elements through the
// mechanism of element substitution groups.
// https://www.w3.org/TR/xmlschema-1/#cElement_Declarations
type Element struct {
	Doc      string
	Name     string
	Wildcard bool
	Type     string
	Abstract bool
	Plural   bool
	Optional bool
	Nillable bool
	Default  string
}

// Attribute declarations provide for: Local validation of attribute
// information item values using a simple type definition; Specifying default
// or fixed values for attribute information items.
// https://www.w3.org/TR/xmlschema-1/structures.html#element-attribute
type Attribute struct {
	Name     string
	Doc      string
	Type     string
	Plural   bool
	Default  string
	Optional bool
}

// ComplexType definitions are identified by their {name} and {target
// namespace}. Except for anonymous complex type definitions (those with no
// {name}), since type definitions (i.e. both simple and complex type
// definitions taken together) must be uniquely identified within an ·XML
// Schema·, no complex type definition can have the same name as another
// simple or complex type definition. Complex type {name}s and {target
// namespace}s are provided for reference from instances, and for use in the
// XML representation of schema components (specifically in <element>). See
// References to schema components across namespaces for the use of component
// identifiers when importing one schema into another.
// https://www.w3.org/TR/xmlschema-1/structures.html#element-complexType
type ComplexType struct {
	Doc            string
	Name           string
	Base           string
	Anonymous      bool
	Elements       []Element
	Attributes     []Attribute
	Groups         []Group
	AttributeGroup []AttributeGroup
	Mixed          bool
}

// Group (model group) definitions are provided primarily for reference from
// the XML Representation of Complex Type Definitions. Thus, model group
// definitions provide a replacement for some uses of XML's parameter entity
// facility.
// https://www.w3.org/TR/xmlschema-1/structures.html#cModel_Group_Definitions
type Group struct {
	Doc      string
	Name     string
	Elements []Element
	Groups   []Group
	Plural   bool
	Ref      string
}

// AttributeGroup definitions do not participate in ·validation· as such, but
// the {attribute uses} and {attribute wildcard} of one or more complex type
// definitions may be constructed in whole or part by reference to an
// attribute group. Thus, attribute group definitions provide a replacement
// for some uses of XML's parameter entity facility. Attribute group
// definitions are provided primarily for reference from the XML
// representation of schema components (see <complexType> and
// <attributeGroup>).
// https://www.w3.org/TR/xmlschema-1/structures.html#Attribute_Group_Definition
type AttributeGroup struct {
	Doc        string
	Name       string
	Ref        string
	Attributes []Attribute
}

// Restriction are used to define acceptable values for XML elements or
// attributes. Restriction on XML elements are called facets.
// https://www.w3.org/TR/xmlschema-1/structures.html#element-restriction
type Restriction struct {
	Doc                  string
	Precision            int
	Enum                 []string
	Min, Max             float64
	MinLength, MaxLength int
	Pattern              *regexp.Regexp
}
