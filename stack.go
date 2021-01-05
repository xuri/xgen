// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import "container/list"

// Stack defined an abstract data type that serves as a collection of elements
type Stack struct {
	list *list.List
}

// NewStack create a new stack
func NewStack() *Stack {
	list := list.New()
	return &Stack{list}
}

// Push a value onto the top of the stack
func (stack *Stack) Push(value interface{}) {
	stack.list.PushBack(value)
}

// Pop the top item of the stack and return it
func (stack *Stack) Pop() interface{} {
	e := stack.list.Back()
	if e != nil {
		stack.list.Remove(e)
		return e.Value
	}
	return nil
}

// Peek view the top item on the stack
func (stack *Stack) Peek() interface{} {
	e := stack.list.Back()
	if e != nil {
		return e.Value
	}
	return nil
}

// Len return the number of items in the stack
func (stack *Stack) Len() int {
	return stack.list.Len()
}

// Empty the stack
func (stack *Stack) Empty() bool {
	return stack.list.Len() == 0
}
