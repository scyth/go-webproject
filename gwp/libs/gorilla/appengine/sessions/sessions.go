// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"time"

	"appengine/datastore"
	"appengine/memcache"

	"code.google.com/p/gorilla/securecookie"
	"code.google.com/p/gorilla/sessions"
)

// DatastoreStore -------------------------------------------------------------

// Session is used to load and save session data in the datastore.
type Session struct {
	Date  time.Time
	Value []byte
}

// NewDatastoreStore returns a new DatastoreStore.
//
// The kind argument is the kind name used to store the session data.
// If empty it will use "Session".
//
// See NewCookieStore() for a description of the other parameters.
func NewDatastoreStore(kind string, keyPairs ...[]byte) *DatastoreStore {
	if kind == "" {
		kind = "Session"
	}
	return &DatastoreStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		kind: kind,
	}
}

// DatastoreStore stores sessions in the App Engine datastore.
type DatastoreStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration
	kind    string
}

// Get returns a session for the given name after adding it to the registry.
//
// See CookieStore.Get().
func (s *DatastoreStore) Get(r *http.Request, name string) (*sessions.Session,
	error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See CookieStore.New().
func (s *DatastoreStore) New(r *http.Request, name string) (*sessions.Session,
	error) {
	session := sessions.NewSession(s, name)
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(r, session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *DatastoreStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	if session.ID == "" {
		session.ID = string(securecookie.GenerateRandomKey(32))
	}
	if err := s.save(r, session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	options := s.Options
	if session.Options != nil {
		options = session.Options
	}
	cookie := &http.Cookie{
		Name:     session.Name(),
		Value:    encoded,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
	http.SetCookie(w, cookie)
	return nil
}

// save writes encoded session.Values to datastore.
func (s *DatastoreStore) save(r *http.Request,
	session *sessions.Session) error {
	if len(session.Values) == 0 {
		// Don't need to write anything.
		return nil
	}
	serialized, err := serialize(session.Values)
	if err != nil {
		return err
	}
	c := newContext(r)
	k := datastore.NewKey(c, s.kind, session.ID, 0, nil)
	k, err = datastore.Put(c, k, &Session{
		Date:  time.Now(),
		Value: serialized,
	})
	if err != nil {
		return err
	}
	return nil
}

// load gets a value from datastore and decodes its content into
// session.Values.
func (s *DatastoreStore) load(r *http.Request,
	session *sessions.Session) error {
	c := newContext(r)
	k := datastore.NewKey(c, s.kind, session.ID, 0, nil)
	entity := Session{}
	if err := datastore.Get(c, k, &entity); err != nil {
		return err
	}
	if err := deserialize(entity.Value, &session.Values); err != nil {
		return err
	}
	return nil
}

// MemcacheStore --------------------------------------------------------------

// NewMemcacheStore returns a new MemcacheStore.
//
// The keyPrefix argument is the prefix used for memcache keys. If empty it
// will use "gorilla.appengine.sessions.".
//
// See NewCookieStore() for a description of the other parameters.
func NewMemcacheStore(keyPrefix string, keyPairs ...[]byte) *MemcacheStore {
	if keyPrefix == "" {
		keyPrefix = "gorilla.appengine.sessions."
	}
	return &MemcacheStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		prefix: keyPrefix,
	}
}

// MemcacheStore stores sessions in the App Engine memcache.
type MemcacheStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration
	prefix  string
}

// Get returns a session for the given name after adding it to the registry.
//
// See CookieStore.Get().
func (s *MemcacheStore) Get(r *http.Request, name string) (*sessions.Session,
	error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// See CookieStore.New().
func (s *MemcacheStore) New(r *http.Request, name string) (*sessions.Session,
	error) {
	session := sessions.NewSession(s, name)
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(r, session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save adds a single session to the response.
func (s *MemcacheStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	if session.ID == "" {
		session.ID = s.prefix + string(securecookie.GenerateRandomKey(32))
	}
	if err := s.save(r, session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	options := s.Options
	if session.Options != nil {
		options = session.Options
	}
	cookie := &http.Cookie{
		Name:     session.Name(),
		Value:    encoded,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
	http.SetCookie(w, cookie)
	return nil
}

// save writes encoded session.Values to memcache.
func (s *MemcacheStore) save(r *http.Request,
	session *sessions.Session) error {
	if len(session.Values) == 0 {
		// Don't need to write anything.
		return nil
	}
	serialized, err := serialize(session.Values)
	if err != nil {
		return err
	}
	err = memcache.Set(newContext(r), &memcache.Item{
		Key:   session.ID,
		Value: serialized,
	})
	if err != nil {
		return err
	}
	return nil
}

// load gets a value from memcache and decodes its content into session.Values.
func (s *MemcacheStore) load(r *http.Request,
	session *sessions.Session) error {
	item, err := memcache.Get(newContext(r), session.ID)
	if err != nil {
		return err
	}
	if err := deserialize(item.Value, &session.Values); err != nil {
		return err
	}
	return nil
}

// Serialization --------------------------------------------------------------

// serialize encodes a value using gob.
func serialize(src interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserialize decodes a value using gob.
func deserialize(src []byte, dst interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(src))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}
