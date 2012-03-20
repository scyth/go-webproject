// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"testing"
)

type keyType int

const (
	key1 keyType = iota
	key2
)

func TestContext(t *testing.T) {
	assertEqual := func(val interface{}, exp interface{}) {
		if val != exp {
			t.Errorf("Expected %v, got %v.", exp, val)
		}
	}

	c := new(Context)
	r, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	// Get()
	assertEqual(c.Get(r, key1), nil)

	// Set()
	c.Set(r, key1, "1")
	assertEqual(c.Get(r, key1), "1")
	assertEqual(len(c.m[r]), 1)

	c.Set(r, key2, "2")
	assertEqual(c.Get(r, key2), "2")
	assertEqual(len(c.m[r]), 2)

	// Delete()
	c.Delete(r, key1)
	assertEqual(c.Get(r, key1), nil)
	assertEqual(len(c.m[r]), 1)

	c.Delete(r, key2)
	assertEqual(c.Get(r, key2), nil)
	assertEqual(len(c.m[r]), 0)

	// Clear()
	c.Clear(r)
	assertEqual(len(c.m), 0)
}
