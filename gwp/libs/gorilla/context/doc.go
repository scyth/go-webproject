// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/context provides a container to store values for a request.

A context stores global variables used during a request. For example, a router
can set variables extracted from the URL and later application handlers can
access those values. There are several others common cases.

The context idea was posted by Brad Fitzpatrick to the go-nuts mailing list:

	http://groups.google.com/group/golang-nuts/msg/e2d679d303aa5d53

Here's the basic usage: first define the keys that you will need. The key
type is interface{} so a key can be of any type that supports equality.
Here we define a key using a custom int type to avoid name collisions:

	package foo

	type contextKey int

	const Key1 contextKey = 0

Then set a variable in the context. Context variables are bound to a
http.Request object, so you need a request instance to set a value:

	context.DefaultContext.Set(request, Key1, "bar")

The application can later access the variable using the same key you provided:

	func MyHandler(w http.ResponseWriter, r *http.Request) {
		// val is "bar".
		val = context.DefaultContext.Get(r, foo.Key1)

		// ...
	}

And that's all about the basic usage. We discuss some other ideas below.

A Context can store any type. To enforce a given type, make the key private
and wrap Get() and Set() to accept and return values of a specific type:

	type contextKey int

	const key1 contextKey = 0

	// GetKey1 returns a value for this package from the request context.
	func GetKey1(request *http.Request) SomeType {
		if rv := context.DefaultContext.Get(request, key1); rv != nil {
			return rv.(SomeType)
		}
		return nil
	}

	// SetKey1 sets a value for this package in the request context.
	func SetKey1(request *http.Request, val SomeType) {
		context.DefaultContext.Set(request, key1, val)
	}

A context must be cleared at the end of a request, to remove all values
that were stored. This can be done in a http.Handler, after a request was
served. Just call Clear() passing the request:

	context.DefaultContext.Clear(request)

The package gorilla/mux clears the default context, so if you are using the
default handler from there you don't need to do anything: context variables
will be deleted at the end of a request.
*/
package context
