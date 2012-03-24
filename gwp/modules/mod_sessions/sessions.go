// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mod_sessions

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"time"
	"github.com/scyth/go-webproject/gwp/libs/gorilla/context"
)

func init() {
	// gob encoding/decoding: register types for sessions and flashes.
	gob.Register(SessionData{})
	gob.Register([]interface{}{})
	// Register the default session store.
	DefaultSessionFactory.SetStore("cookie", new(CookieSessionStore))
}

var (
	DefaultSessionKey = "s"
	DefaultStoreKey   = "cookie"
	DefaultFlashesKey = "_flash"
	// All errors.
	ErrEncoding       = errors.New("The value could not be encoded.")
	ErrDecoding       = errors.New("The value could not be decoded.")
	ErrAuthentication = errors.New("The value could not be verified using HMAC.")
	ErrDecryption     = errors.New("The value could not be decrypted.")
	ErrMaxLength      = errors.New("The value exceeds the maximum allowed length.")
	ErrBadTimestamp   = errors.New("Invalid timestamp.")
	ErrNewTimestamp   = errors.New("The value has a newer timestamp than expected.")
	ErrOldTimestamp   = errors.New("The value has an expired timestamp.")
	ErrMissingHash    = errors.New("A hash is required to create and verify values using HMAC.")
	ErrMissingHashKey = errors.New("Authentication secret can't be nil.")
	ErrNoSession      = errors.New("No session found for the given key.")
	ErrNoFlashes      = errors.New("No flashes found for the given key.")
	ErrNoStore        = errors.New("No store found for the given key.")
	ErrStoreMismatch  = errors.New("A session with the given key already exists using a different store.")
	ErrBadIdLength    = errors.New("Session id length must be greater than zero.")
)

// The type used to store session values.
type SessionData map[string]interface{}

// SessionConfig stores configuration for each session.
//
// Fields are a subset of http.Cookie fields.
type SessionConfig struct {
	Path   string
	Domain string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

// SessionInfo stores internal references for a given session.
type SessionInfo struct {
	Data   SessionData
	Store  SessionStore
	Config SessionConfig
}

// ----------------------------------------------------------------------------
// Convenience functions
// ----------------------------------------------------------------------------

// Session returns a session for the current request.
//
// The variadic arguments are optional: (sessionKey, storeKey). They are used
// to load a different session key or use a session store other than the
// default one. If not defined or empty the defaults are used.
func Session(r *http.Request, vars ...string) (SessionData, error) {
	return DefaultSessionFactory.Session(r, vars...)
}

// Flashes returns an array of flash messages, if any.
//
// The variadic arguments are optional: (flashKey, sessionKey, storeKey).
// If not defined or empty the default values are used.
func Flashes(r *http.Request, vars ...string) ([]interface{}, error) {
	return DefaultSessionFactory.Flashes(r, vars...)
}

// AddFlash adds a flash message.
//
// The variadic arguments are optional: (flashKey, sessionKey, storeKey).
// If not defined or empty the default values are used.
func AddFlash(r *http.Request, value interface{},
	vars ...string) (bool, error) {
	return DefaultSessionFactory.AddFlash(r, value, vars...)
}

// Config returns the configuration for a given session.
//
// The key argument is optional; if not set it'll use the default session key.
func Config(r *http.Request, key ...string) (*SessionConfig, error) {
	return DefaultSessionFactory.Config(r, key...)
}

// Save saves all sessions accessed during the request.
func Save(r *http.Request, w http.ResponseWriter) []error {
	return DefaultSessionFactory.Save(r, w)
}

// Store returns a session store for the given key.
func Store(key string) (SessionStore, error) {
	return DefaultSessionFactory.Store(key)
}

// SetStore registers a session store for the given key.
func SetStore(key string, store SessionStore) {
	DefaultSessionFactory.SetStore(key, store)
}

// SetStoreKeys defines authentication and encryption keys for the given store.
//
// See SessionFactory.SetStoreKeys.
func SetStoreKeys(key string, pairs ...[]byte) (bool, error) {
	return DefaultSessionFactory.SetStoreKeys(key, pairs...)
}

// DefaultConfig returns the default session configuration used by the factory.
func DefaultConfig() *SessionConfig {
	return DefaultSessionFactory.DefaultConfig()
}

// SetDefaultConfig sets the default session configuration used by the factory.
func SetDefaultConfig(config *SessionConfig) {
	DefaultSessionFactory.SetDefaultConfig(config)
}

// ----------------------------------------------------------------------------
// SessionFactory
// ----------------------------------------------------------------------------

// DefaultSessionFactory is the default factory for session requests.
var DefaultSessionFactory = new(SessionFactory)

// DefaultSessionConfig is the session configuration used when none is set.
var DefaultSessionConfig = &SessionConfig{
	Path:     "/",
	Domain:   "",
	MaxAge:   86400 * 30,
	Secure:   false,
	HttpOnly: false,
}

// SessionFactory registers configuration and stores available for use.
type SessionFactory struct {
	stores        map[string]SessionStore
	defaultConfig *SessionConfig
}

// Store returns a session store for the given key.
func (f *SessionFactory) Store(key string) (SessionStore, error) {
	store, ok := f.stores[key]
	if !ok {
		return nil, ErrNoStore
	}
	return store, nil
}

// SetStore registers a session store for the given key.
func (f *SessionFactory) SetStore(key string, store SessionStore) {
	if f.stores == nil {
		f.stores = make(map[string]SessionStore)
	}
	f.stores[key] = store
}

// SetStoreKeys defines authentication and encryption keys for the given store.
//
// This is a convenience to set secret keys for the available stores.
// It sets authentication hashes using HMAC-SHA-256 and encryption blocks
// using AES. For custom hash or encryption methods, call
// SessionStore.SetEncoders() directly.
//
// A store must be registered using the given key before this is called.
//
// Keys are defined in pairs: one for authentication and the other for
// encryption. The encryption key can be set to nil or omitted in the last
// pair, but the authentication key is required in all pairs.
//
// Multiple pairs are accepted to allow key rotation, but the common case is
// to set a single authentication key and optionally an encryption key.
//
// The encryption key, if set, must be either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256 modes.
func (f *SessionFactory) SetStoreKeys(key string,
	pairs ...[]byte) (bool, error) {
	store, err := f.Store(key)
	if err != nil {
		return false, err
	}
	var b cipher.Block
	size := len(pairs)
	encoders := make([]SessionEncoder, size/2+size%2)
	for i := 0; i < size; i += 2 {
		if pairs[i] == nil || len(pairs[i]) == 0 {
			return false, ErrMissingHashKey
		}
		if size <= i+1 || pairs[i+1] == nil {
			b = nil
		} else {
			b, err = aes.NewCipher(pairs[i+1])
			if err != nil {
				return false, err
			}
		}
		encoders[i/2] = &Encoder{
			Hash:      hmac.New(sha256.New, pairs[i]),
			Block:     b,
			MaxAge:    86400 * 30,
			MaxLength: 4096,
		}
	}
	// Set the new encoders.
	store.SetEncoders(encoders...)
	return true, nil
}

// Session returns a session for the current request.
//
// The variadic arguments are optional: (sessionKey, storeKey). They are used
// to load a different session key or use a session store other than the
// default one. If not defined or empty the defaults are used.
func (f *SessionFactory) Session(r *http.Request,
	vars ...string) (SessionData, error) {
	return getRequestSessions(f, r).Session(vars...)
}

// Flashes returns an array of flash messages, if any.
//
// The variadic arguments are optional: (flashKey, sessionKey, storeKey).
// If not defined or empty the default values are used.
func (f *SessionFactory) Flashes(r *http.Request,
	vars ...string) ([]interface{}, error) {
	key, newvars := flashKey(vars...)
	session, err := f.Session(r, newvars...)
	if err != nil {
		return nil, err
	}
	if flashes, ok := session[key]; ok {
		// Drop the flashes and return it.
		delete(session, key)
		return flashes.([]interface{}), nil
	}
	return nil, ErrNoFlashes
}

// AddFlash adds a flash message.
//
// The variadic arguments are optional: (flashKey, sessionKey, storeKey).
// If not defined or empty the default values are used.
func (f *SessionFactory) AddFlash(r *http.Request, value interface{},
	vars ...string) (bool, error) {
	key, newvars := flashKey(vars...)
	session, err := f.Session(r, newvars...)
	if err != nil {
		return false, err
	}
	var flashes []interface{}
	if v, ok := session[key]; ok {
		flashes = v.([]interface{})
	} else {
		flashes = make([]interface{}, 0)
	}
	session[key] = append(flashes, value)
	return true, nil
}

// Config returns the configuration for a given session.
//
// The key argument is optional; if not set it'll use the default session key.
func (f *SessionFactory) Config(r *http.Request,
	key ...string) (*SessionConfig, error) {
	return getRequestSessions(f, r).Config(key...)
}

// Save saves all sessions accessed during the request.
func (f *SessionFactory) Save(r *http.Request,
	w http.ResponseWriter) []error {
	return getRequestSessions(f, r).Save(w)
}

// DefaultConfig returns the default session configuration used by the factory.
func (f *SessionFactory) DefaultConfig() *SessionConfig {
	if f.defaultConfig == nil {
		f.defaultConfig = DefaultSessionConfig
	}
	return f.defaultConfig
}

// SetDefaultConfig sets the default session configuration used by the factory.
func (f *SessionFactory) SetDefaultConfig(config *SessionConfig) {
	f.defaultConfig = config
}

// defaultConfigValue returns a copy of the default configuration.
func (f *SessionFactory) defaultConfigValue() SessionConfig {
	d := f.DefaultConfig()
	return SessionConfig{
		Path:     d.Path,
		Domain:   d.Domain,
		MaxAge:   d.MaxAge,
		Secure:   d.Secure,
		HttpOnly: d.HttpOnly,
	}
}

// flashKey extracts a flashes key from variadic arguments.
func flashKey(vars ...string) (string, []string) {
	key := DefaultFlashesKey
	if len(vars) > 0 {
		if vars[0] != "" {
			key = vars[0]
		}
		vars = vars[1:]
	}
	return key, vars
}

// ----------------------------------------------------------------------------
// Context
// ----------------------------------------------------------------------------

//  The key type is interface{} so a key can be of any type that supports equality.
//  Here we define a key using a custom int type to avoid name collisions:

type contextKey int
const key1 contextKey = 0

// getRequestSessions returns a sessions container for a single request.
func getRequestSessions(f *SessionFactory, r *http.Request) *requestSessions {

	var s *requestSessions
	rv := context.DefaultContext.Get(r, key1)
	if rv != nil {
		s = rv.(*requestSessions)
	} else {
		s = &requestSessions{factory: f, request: r}
		context.DefaultContext.Set(r, key1, s)
	}
	return s
}

// ----------------------------------------------------------------------------
// requestSessions
// ----------------------------------------------------------------------------

// requestSessions stores sessions in use for a given request.
type requestSessions struct {
	factory  *SessionFactory
	request  *http.Request
	sessions map[string]SessionInfo
}

// Session returns a session given its key and store.
//
// The variadic arguments are optional: (sessionKey, storeKey). They are used
// to load a different session key or use a session store other than the
// default one. If not defined or empty the defaults are used.
func (s *requestSessions) Session(vars ...string) (SessionData, error) {
	sessionKey, storeKey := sessionKeys(vars...)
	// Get the requested store.
	store, err := s.factory.Store(storeKey)
	if err != nil {
		return nil, err
	}
	if s.sessions == nil {
		s.sessions = make(map[string]SessionInfo)
	}
	// See if there's an existing session with the given key/store.
	info, ok := s.sessions[sessionKey]
	if ok {
		if store != info.Store {
			// Store should match.
			return nil, ErrStoreMismatch
		}
		return info.Data, nil
	}
	// Load a new session.
	info = SessionInfo{
		Store:  store,
		Config: s.factory.defaultConfigValue(),
	}
	store.Load(s.request, sessionKey, &info)
	s.sessions[sessionKey] = info
	return info.Data, nil
}

// Config returns the configuration for a given session.
//
// The key argument is optional; if not set it'll use the default session key.
func (s *requestSessions) Config(key ...string) (*SessionConfig, error) {
	sessionKey, _ := sessionKeys(key...)
	if info, ok := s.sessions[sessionKey]; ok {
		return &info.Config, nil
	}
	return nil, ErrNoSession
}

// Save saves all sessions accessed during the request.
func (s *requestSessions) Save(w http.ResponseWriter) []error {
	var err error
	var ok bool
	var errors []error
	for key, info := range s.sessions {
		if ok, err = info.Store.Save(s.request, w, key, &info); !ok {
			if errors == nil {
				errors = []error{err}
			} else {
				errors = append(errors, err)
			}
		}
	}
	return errors
}

// sessionKeys extracts session/store keys from variadic arguments.
func sessionKeys(vars ...string) (string, string) {
	sessionKey := DefaultSessionKey
	storeKey := DefaultStoreKey
	if len(vars) > 0 && vars[0] != "" {
		sessionKey = vars[0]
	}
	if len(vars) > 1 && vars[1] != "" {
		storeKey = vars[1]
	}
	return sessionKey, storeKey
}

// ----------------------------------------------------------------------------
// SessionStore
// ----------------------------------------------------------------------------

// SessionStore defines an interface for session stores.
type SessionStore interface {
	Load(r *http.Request, key string, info *SessionInfo)
	Save(r *http.Request, w http.ResponseWriter, key string, info *SessionInfo) (bool, error)
	Init(r *http.Request, w http.ResponseWriter, key string, info *SessionInfo) (bool, error)
	Encoders() []SessionEncoder
	SetEncoders(encoders ...SessionEncoder)
}

// ----------------------------------------------------------------------------
// CookieSessionStore
// ----------------------------------------------------------------------------

// CookieSessionStore is the default session store.
//
// It stores the session data in authenticated and, optionally, encrypted
// cookies.
type CookieSessionStore struct {
	// List of encoders registered for this store.
	encoders []SessionEncoder
}

// Load loads a session for the given key.
func (s *CookieSessionStore) Load(r *http.Request, key string,
	info *SessionInfo) {
	info.Data = GetCookie(s, r, key)
}

// Save saves the session in the response.
func (s *CookieSessionStore) Save(r *http.Request, w http.ResponseWriter,
	key string, info *SessionInfo) (bool, error) {
	return SetCookie(s, w, key, info)
}

// Encoders returns the encoders for this store.
func (s *CookieSessionStore) Encoders() []SessionEncoder {
	return s.encoders
}

// SetEncoders sets a group of encoders in the store.
func (s *CookieSessionStore) SetEncoders(encoders ...SessionEncoder) {
	s.encoders = encoders
}

// ----------------------------------------------------------------------------
// Utilities for custom session stores
// ----------------------------------------------------------------------------

// GenerateSessionId generates a random session id with the given length.
func GenerateSessionId(length int) (string, error) {
	if length <= 0 {
		return "", ErrBadIdLength
	}
	id := make([]byte, length)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", id), nil
}

// GetCookie returns the contents from a session cookie.
//
// If the session is invalid, it will return an empty SessionData.
func GetCookie(s SessionStore, r *http.Request, key string) SessionData {
	if cookie, err := r.Cookie(key); err == nil {
		if data, err2 := Decode(s, key, cookie.Value); err2 == nil {
			return data
		}
	}
	return SessionData{}
}

// SetCookie sets a session cookie using the user-defined configuration.
//
// Custom backends will only store a session id in the cookie.
func SetCookie(s SessionStore, w http.ResponseWriter, key string,
	info *SessionInfo) (bool, error) {
	encoded, err := Encode(s, key, info.Data)
	if err != nil {
		return false, err
	}
	cookie := &http.Cookie{
		Name:     key,
		Value:    encoded,
		Path:     info.Config.Path,
		Domain:   info.Config.Domain,
		MaxAge:   info.Config.MaxAge,
		Secure:   info.Config.Secure,
		HttpOnly: info.Config.HttpOnly,
	}
	http.SetCookie(w, cookie)
	return true, nil
}

// Encode encodes a session value for a session store.
func Encode(s SessionStore, key string, value SessionData) (string, error) {
	encoders := s.Encoders()
	if encoders != nil {
		var encoded string
		var err error
		for _, encoder := range encoders {
			encoded, err = encoder.Encode(key, value)
			if err == nil {
				return encoded, nil
			}
		}
	}
	return "", ErrEncoding
}

// Decode decodes a session value for a session store.
func Decode(s SessionStore, key, value string) (SessionData, error) {
	encoders := s.Encoders()
	if encoders != nil {
		var decoded SessionData
		var err error
		for _, encoder := range encoders {
			decoded, err = encoder.Decode(key, value)
			if err == nil {
				return decoded, nil
			}
		}
	}
	return nil, ErrDecoding
}

// SerializeSessionData serializes a session value using gob.
func SerializeSessionData(session SessionData) ([]byte, error) {
	return serialize(session)
}

// DeserializeSessionData deserializes a session value using gob.
func DeserializeSessionData(value []byte) (data SessionData, err error) {
	return deserialize(value)
}

// ----------------------------------------------------------------------------
// SessionEncoder
// ----------------------------------------------------------------------------

// SessionEncoder defines an interface to encode and decode session values.
type SessionEncoder interface {
	Encode(key string, value SessionData) (string, error)
	Decode(key, value string) (SessionData, error)
}

// ----------------------------------------------------------------------------
// Encoder
// ----------------------------------------------------------------------------

// Encoder encodes and decodes session values.
//
// It is a default SessionEncoder implementation available for all session
// stores. It performs up to four operations in both ways:
//
// 1. Serialization: encodes to and from gob.
//
// 2. Encryption (optional): if the Block field is set, it is used to encrypt
// and decrypt the value in CTR mode.
//
// 3. Authentication: creates and verifies HMACs. The Hash field is required:
// if not set, sessions can't be read or written.
//
// 4. Encoding: converts to and from a format suitable for cookie transmition.
//
// Multiple encoders can be added to a session store to allow secret keys
// rotation.
type Encoder struct {
	// Required, used for authentication.
	// Set it to, e.g.: hmac.NewSHA256([]byte("very-secret-key"))
	Hash hash.Hash
	// Optional, used for encryption.
	// Set it to, e.g.: aes.NewCipher([]byte("16-length-secret-key"))
	Block cipher.Block
	// Optional, to restrict minimum age, in seconds, for the timestamp value.
	// Set it to 0 for no restriction.
	MinAge int64
	// Optional, to restrict maximum age, in seconds, for the timestamp value.
	// Set it to 0 for no restriction.
	MaxAge int64
	// Optional, to restrict length of values to be decoded.
	// Set it to, e.g.: 1024 (conservative) or 4096 (maximum cookie size).
	MaxLength int
	// For testing purposes, the function that returns the current timestamp.
	// If not set, it will use time.UTC().Seconds().
	TimeFunc func() int64
}

// Encode encodes a session value.
//
// It serializes, optionally encrypts, creates a message authentication code
// and finally encodes the value in a format suitable for cookie transmition.
func (s *Encoder) Encode(key string, value SessionData) (rv string, err error) {
	// Hash is required.
	if s.Hash == nil {
		err = ErrMissingHash
		return
	}
	var b []byte

	// 1. Serialize.
	b, err = serialize(value)
	if err != nil {
		return
	}

	// 2. Encrypt (optional).
	if s.Block != nil {
		b, err = encrypt(s.Block, b)
		if err != nil {
			return
		}
		// Encode because pipes would break HMAC verification.
		b = encode(b)

	}

	// 3. Create hash.
	b = createHmac(s.Hash, key, b, s.timestamp())

	// 4. Encode.
	rv = string(encode(b))
	return
}

// Decode decodes a session value.
//
// It decodes, verifies a message authentication code, optionally decrypts and
// finally deserializes the value.
func (s *Encoder) Decode(key, value string) (SessionData, error) {
	// Hash is required.
	if s.Hash == nil {
		return nil, ErrMissingHash
	}

	// Check max length.
	if s.MaxLength != 0 && len(value) > s.MaxLength {
		return nil, ErrMaxLength
	}

	// 1. Decode.
	rv, err := decode([]byte(value))
	if err != nil {
		return nil, err
	}

	// 2. Verify hash.
	rv, err = verifyHmac(s.Hash, key, rv, s.timestamp(), s.MinAge, s.MaxAge)
	if err != nil {
		return nil, err
	}

	// 3. Decrypt (optional).
	if s.Block != nil {
		rv, err = decode(rv)
		if err != nil {
			return nil, err
		}
		rv, err = decrypt(s.Block, rv)
		if err != nil {
			return nil, err
		}
	}

	// 4. Deserialize.
	var data SessionData
	data, err = deserialize(rv)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// timestamp returns the current timestamp, in seconds.
//
// For testing purposes, the function that generates the timestamp can be
// overridden. If TimeFunc is not set, it will return time.UTC().Seconds().
func (s *Encoder) timestamp() int64 {
	if s.TimeFunc == nil {
		return time.Now().UTC().Unix()
	}
	return s.TimeFunc()
}

// Serialization --------------------------------------------------------------

// serialize encodes a session value using gob.
func serialize(session SessionData) ([]byte, error) {
	b := bytes.NewBuffer(nil)
	e := gob.NewEncoder(b)
	if err := e.Encode(session); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// deserialize decodes a session value using gob.
func deserialize(value []byte) (SessionData, error) {
	var session SessionData
	b := bytes.NewBuffer(value)
	d := gob.NewDecoder(b)
	if err := d.Decode(&session); err != nil {
		return nil, err
	}
	return session, nil
}

// Encryption -----------------------------------------------------------------

// encrypt encrypts a value using the given Block in CTR mode.
//
// A random initialization vector is generated and prepended to the resulting
// ciphertext to be available for decryption. Also, a random salt with the
// length of the block size is prepended to the value before encryption.
func encrypt(block cipher.Block, value []byte) (rv []byte, err error) {
	// Recover in case block has an invalid key.
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	size := block.BlockSize()
	// Generate an initialization vector suitable for encryption.
	// http://en.wikipedia.org/wiki/Block_cipher_modes_of_operation#Initialization_vector_.28IV.29
	iv := make([]byte, size)
	if _, err = rand.Read(iv); err != nil {
		return
	}
	// Create a salt.
	salt := make([]byte, size)
	if _, err = rand.Read(salt); err != nil {
		return
	}
	value = append(salt, value...)
	// Encrypt it.
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(value, value)
	// Return iv + ciphertext.
	rv = append(iv, value...)
	return
}

// decrypt decrypts a value using the given Block in CTR mode.
//
// The value to be decrypted must have a length greater than the block size,
// because the initialization vector is expected to prepend it. Also, a salt
// with the length of the block size is expected to prepend the plain value.
func decrypt(block cipher.Block, value []byte) (b []byte, err error) {
	// Recover in case block has an invalid key.
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	size := block.BlockSize()
	if len(value) > size {
		// Extract iv.
		iv := value[:size]
		// Extract ciphertext.
		value = value[size:]
		// Decrypt it.
		stream := cipher.NewCTR(block, iv)
		stream.XORKeyStream(value, value)
		if len(value) > size {
			// Return value without the salt.
			b = value[size:]
			return
		}
	}
	err = ErrDecryption
	return
}

// Authentication -------------------------------------------------------------

// createHmac creates a message authentication code (MAC).
//
// It returns the concatenation of "value|timestamp|message".
func createHmac(h hash.Hash, key string, value []byte,
	timestamp int64) []byte {
	msg := mac(h, key, value, timestamp)
	return []byte(fmt.Sprintf("%s|%d|%s", value, timestamp, msg))
}

// verifyHmac verifies that a message authentication code (MAC) is valid.
//
// The provided source bytes must be in the form "value|timestamp|message".
func verifyHmac(h hash.Hash, key string, value []byte, timestamp, minAge,
	maxAge int64) ([]byte, error) {
	parts := bytes.SplitN(value, []byte("|"), 3)
	if len(parts) != 3 {
		return nil, ErrAuthentication
	}
	rv := parts[0]
	tst, _ := strconv.ParseInt(string(parts[1]), 10, 64)
	msg := parts[2]
	if tst == 0 {
		return nil, ErrBadTimestamp
	}
	if minAge != 0 && tst > timestamp-minAge {
		return nil, ErrNewTimestamp
	}
	if maxAge != 0 && tst < timestamp-maxAge {
		return nil, ErrOldTimestamp
	}
	// There are several other operations being done by the Encoder so not
	// sure if ConstantTimeCompare() is worth at all.
	msg2 := mac(h, key, rv, tst)
	if len(msg) != len(msg2) || subtle.ConstantTimeCompare(msg, msg2) != 1 {
		return nil, ErrAuthentication
	}
	return rv, nil
}

// mac generates a message authentication code (MAC).
//
// The message is created with the concatenation of "key|value|timestamp".
func mac(h hash.Hash, key string, value []byte, timestamp int64) []byte {
	h.Reset()
	h.Write([]byte(fmt.Sprintf("%s|%s|%d", key, value, timestamp)))
	return h.Sum([]byte(fmt.Sprintf("%s|%s|%d", key, value, timestamp)))
}

// Encoding -------------------------------------------------------------------

// encode encodes a value to a format suitable for cookie transmission.
func encode(value []byte) []byte {
	encoded := make([]byte, base64.URLEncoding.EncodedLen(len(value)))
	base64.URLEncoding.Encode(encoded, value)
	return encoded
}

// decode decodes a value received as a session cookie.
func decode(value []byte) ([]byte, error) {
	decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
	b, err := base64.URLEncoding.Decode(decoded, value)
	if err != nil {
		return nil, err
	}
	return decoded[:b], nil
}
