// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

// Entities with more than this many indexed properties will not be saved.
const maxIndexedProperties = 5000

// []byte fields more than 1 megabyte long will not be loaded or saved.
const maxBlobLen = 1 << 20

// Property is a name/value pair plus some metadata. A datastore entity's
// contents are loaded and saved as a sequence of Properties. An entity can
// have multiple Properties with the same name, provided that p.Multiple is
// true on all of that entity's Properties with that name.
type Property struct {
	// Name is the property name.
	Name string
	// Value is the property value. The valid types are:
	//	- int64
	//	- bool
	//	- string
	//	- float64
	//	- *Key
	//	- time.Time
	//	- appengine.BlobKey
	//	- []byte (up to 1 megabyte in length)
	// This set is smaller than the set of valid struct field types that the
	// datastore can load and save. A Property Value cannot be a slice (apart
	// from []byte); use multiple Properties instead. Also, a Value's type
	// must be explicitly on the list above; it is not sufficient for the
	// underlying type to be on that list. For example, a Value of "type
	// myInt64 int64" is invalid. Smaller-width integers and floats are also
	// invalid. Again, this is more restrictive than the set of valid struct
	// field types.
	Value interface{}
	// NoIndex is whether the datastore cannot index this property.
	NoIndex bool
	// Multiple is whether the entity can have multiple properties with
	// the same name. Even if a particular instance only has one property with
	// a certain name, Multiple should be true if a struct would best represent
	// it as a field of type []T instead of type T.
	Multiple bool
}

// PropertyLoadSaver can be converted from and to a sequence of Properties.
// Load should drain the channel until closed, even if an error occurred.
// Save should close the channel when done, even if an error occurred.
type PropertyLoadSaver interface {
	Load(<-chan Property) error
	Save(chan<- Property) error
}

// PropertyList converts a []Property to implement PropertyLoadSaver.
type PropertyList []Property

var (
	typeOfPropertyLoadSaver = reflect.TypeOf((*PropertyLoadSaver)(nil)).Elem()
	typeOfPropertyList      = reflect.TypeOf(PropertyList(nil))
)

// Load loads all of c's properties into l.
// It does not first reset *l to an empty slice.
func (l *PropertyList) Load(c <-chan Property) error {
	for p := range c {
		*l = append(*l, p)
	}
	return nil
}

// Save saves all of l's properties to c.
func (l *PropertyList) Save(c chan<- Property) error {
	for _, p := range *l {
		c <- p
	}
	close(c)
	return nil
}

// validPropertyName returns whether s is a valid Go field name.
func validPropertyName(s string) bool {
	if s == "" {
		return false
	}
	first := true
	for _, c := range s {
		if first {
			first = false
			if c != '_' && !unicode.IsLetter(c) {
				return false
			}
		} else {
			if c != '_' && !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

// structTag is the parsed `datastore:"name,options"` tag of a struct field.
// If a field has no tag, or the tag has an empty name, then the structTag's
// name is just the field name. A "-" name means that the datastore ignores
// that field.
type structTag struct {
	name    string
	noIndex bool
}

// structCodec describes how to convert a struct to and from a sequence of
// properties.
type structCodec struct {
	// byIndex gives the structTag for the i'th field.
	byIndex []structTag
	// byName gives the field index for the structTag with the given name.
	byName map[string]int
}

// structCodecs collects the structCodecs that have already been calculated.
var (
	structCodecsMutex sync.Mutex
	structCodecs      = make(map[reflect.Type]structCodec)
)

// getStructCodec returns the structCodec for the given struct type.
func getStructCodec(t reflect.Type) (structCodec, error) {
	structCodecsMutex.Lock()
	defer structCodecsMutex.Unlock()
	c, ok := structCodecs[t]
	if ok {
		return c, nil
	}
	c.byIndex = make([]structTag, t.NumField())
	c.byName = make(map[string]int)
	for i := range c.byIndex {
		f := t.Field(i)
		name, opts := f.Tag.Get("datastore"), ""
		if i := strings.Index(name, ","); i != -1 {
			name, opts = name[:i], name[i+1:]
		}
		if name == "" {
			name = f.Name
		} else if name == "-" {
			c.byIndex[i] = structTag{name: name}
			continue
		} else if !validPropertyName(name) {
			return structCodec{}, fmt.Errorf("datastore: struct tag has invalid property name: %q", name)
		} else if _, ok := c.byName[name]; ok {
			return structCodec{}, fmt.Errorf("datastore: struct tag has repeated property name: %q", name)
		}
		c.byIndex[i] = structTag{
			name:    name,
			noIndex: opts == "noindex",
		}
		c.byName[name] = i
	}
	structCodecs[t] = c
	return c, nil
}

// structPLS adapts a struct to be a PropertyLoadSaver.
type structPLS struct {
	v     reflect.Value
	codec structCodec
}

// newStructPLS returns a PropertyLoadSaver for the struct pointer p.
func newStructPLS(p interface{}) (PropertyLoadSaver, error) {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidEntityType
	}
	v = v.Elem()
	codec, err := getStructCodec(v.Type())
	if err != nil {
		return nil, err
	}
	return structPLS{v, codec}, nil
}

// LoadStruct loads the properties from c to dst, reading from c until closed.
// dst must be a struct pointer.
func LoadStruct(dst interface{}, c <-chan Property) error {
	x, err := newStructPLS(dst)
	if err != nil {
		for _ = range c {
			// Drain the channel.
		}
		return err
	}
	return x.Load(c)
}

// SaveStruct saves the properties from src to c, closing c when done.
// src must be a struct pointer.
func SaveStruct(src interface{}, c chan<- Property) error {
	x, err := newStructPLS(src)
	if err != nil {
		close(c)
		return err
	}
	return x.Save(c)
}
