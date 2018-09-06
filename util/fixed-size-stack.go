/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package util

// FixedSizeStack is a stack that works like a regular stack
// (LIFO access: last element that is added is pulled out first)
// but it keeps only a predefined number of items.
// When the maximum size is reached, the oldest item is purged
// from the back of the stack.

import (
	"container/list"
	"sync"
)

// The stack is implemented using a double-linked list from
// Go standard library
type FixedSizeStack struct {
	list     *list.List
	maxItems int
	mux      sync.Mutex
}

// NewFixedSizeStack returns a new stack with a predefined
// size
func NewFixedSizeStack(wanted_size int) (fss FixedSizeStack) {
	fss.maxItems = wanted_size
	fss.list = list.New()
	return
}

// The length of the stack is that of the underlying list
func (fss *FixedSizeStack) Len() int {
	// Locks so that it is safe to get the length from concurrent goroutines
	fss.mux.Lock()
	length := fss.list.Len()
	fss.mux.Unlock()
	return length
}

// Push() inserts an item into the stack.
// The item can be of any type
func (fss *FixedSizeStack) Push(item interface{}) {
	// Locks so that it is safe to push from concurrent goroutines
	fss.mux.Lock()
	// Purges the oldest element if size exceeds
	// maximum allowed storage
	if fss.list.Len() >= fss.maxItems {
		// removes item from the back
		fss.list.Remove(fss.list.Back())
	}
	fss.list.PushFront(item)
	fss.mux.Unlock()
}

// Pop() returns the object stored as .Value in the top list element
// Client calls will need to cast the object to the expected type.
// e.g.:
//     type MyType struct { ... }
//     var lastOne MyType
//     lastOne = fss.Pop().(MyType)
func (fss *FixedSizeStack) Pop() interface{} {
	// Locks so that it is safe to pop from concurrent goroutines
	fss.mux.Lock()
	latest := fss.list.Front().Value
	// After extracting the item, we remove the first
	// list element
	fss.list.Remove(fss.list.Front())
	fss.mux.Unlock()
	return latest
}
