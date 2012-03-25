// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !appengine

package sessions

import (
	"net/http"

	"appengine"

	"gae-go-testing.googlecode.com/git/appenginetesting"
)

// Testing hack :( ------------------------------------------------------------

// context is a testing hack. We don't have a good testing story in App Engine
// so we need this kind of stuff.
var context appengine.Context

func newContext(r *http.Request) appengine.Context {
	if appengine.IsDevAppServer() && r.Header.Get("App-Testing") != "" {
		if context == nil {
			var err error
			if context, err = appenginetesting.NewContext(nil); err != nil {
				panic(err)
			}
		}
		return context
	}
	return appengine.NewContext(r)
}

// closeTestingContext is part of a hack to make packages testable in
// App Engine. :(
func closeTestingContext() {
	if context != nil {
		context.(*appenginetesting.Context).Close()
		context = nil
	}
}
