// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"sync"
)

// Original implementation by Brad Fitzpatrick:
// http://groups.google.com/group/golang-nuts/msg/e2d679d303aa5d53

// DefaultContext is a default context instance.
var DefaultContext = new(Context)

// Context stores values for requests.
type Context struct {
	l sync.Mutex
	m map[*http.Request]map[interface{}]interface{}
}

// Set stores a value for a given key in a given request.
func (c *Context) Set(req *http.Request, key, val interface{}) {
	c.l.Lock()
	defer c.l.Unlock()
	if c.m == nil {
		c.m = make(map[*http.Request]map[interface{}]interface{})
	}
	if c.m[req] == nil {
		c.m[req] = make(map[interface{}]interface{})
	}
	c.m[req][key] = val
}

// Get returns a value registered for a given key in a given request.
func (c *Context) Get(req *http.Request, key interface{}) interface{} {
	c.l.Lock()
	defer c.l.Unlock()
	if c.m != nil && c.m[req] != nil {
		return c.m[req][key]
	}
	return nil
}

// Delete removes the value for a given key in a given request.
func (c *Context) Delete(req *http.Request, key interface{}) {
	c.l.Lock()
	defer c.l.Unlock()
	if c.m != nil && c.m[req] != nil {
		delete(c.m[req], key)
	}
}

// Clear removes all values for a given request.
func (c *Context) Clear(req *http.Request) {
	c.l.Lock()
	defer c.l.Unlock()
	if c.m != nil {
		delete(c.m, req)
	}
}
