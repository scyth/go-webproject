// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/appengine/sessions implements session stores for Google App
Engine's datastore and memcache.

Usage is the same as described in gorilla/sessions documentation:

	http://gorilla-web.appspot.com/pkg/gorilla/sessions

...but you'll use the DatastoreStore or MemcacheStore to load and save your
sessions. Let's initialize both:

	var dStore = sessions.NewDatastoreStore("", []byte("very-secret"))
	var mStore = sessions.NewMemcacheStore("", []byte("a-lot-secret"))

After this, call the appropriate store to retrieve a session, and then call
Save() on the session to save it:

	func MyHandler(w http.ResponseWriter, r *http.Request) {
		// Get a session. We're ignoring the error resulted from decoding an
		// existing session: Get() always returns a session, even if empty.
		session, _ := dStore.Get(r, "session-name")
		// Set some session values.
		session.Values["foo"] = "bar"
		session.Values[42] = 43
		// Save it.
		session.Save(r, w)
	}

Check the sessions package documentation for more details about other features.
*/
package sessions
