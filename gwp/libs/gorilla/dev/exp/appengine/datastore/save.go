// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"

	"appengine"
	"code.google.com/p/goprotobuf/proto"

	pb "appengine_internal/datastore"
)

var (
	minTime = time.Unix(0, 0)
	maxTime = time.Unix(int64(math.MaxInt64)/1e6, (int64(math.MaxInt64)%1e6)*1e3)
)

// valueToProto converts a named value to a newly allocated Property.
// The returned error string is empty on success.
func valueToProto(name string, value interface{}, multiple bool) (p *pb.Property, err error) {
	var (
		pv          pb.PropertyValue
		unsupported bool
	)
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Invalid:
		// No-op.
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		pv.Int64Value = proto.Int64(v.Int())
	case reflect.Bool:
		pv.BooleanValue = proto.Bool(v.Bool())
	case reflect.String:
		pv.StringValue = proto.String(v.String())
	case reflect.Float32, reflect.Float64:
		pv.DoubleValue = proto.Float64(v.Float())
	case reflect.Ptr:
		if k, ok := v.Interface().(*Key); ok {
			if k != nil {
				pv.Referencevalue = keyToReferenceValue(k)
			}
		} else {
			unsupported = true
		}
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			if t.Before(minTime) || t.After(maxTime) {
				return nil, errors.New("time value out of range")
			}
			pv.Int64Value = proto.Int64(t.UnixNano() / 1e3)
		} else {
			unsupported = true
		}
	case reflect.Slice:
		if b, ok := v.Interface().([]byte); ok {
			pv.StringValue = proto.String(string(b))
		} else {
			// nvToProto should already catch slice values.
			// If we get here, we have a slice of slice values.
			unsupported = true
		}
	default:
		unsupported = true
	}
	if unsupported {
		return nil, fmt.Errorf("unsupported datastore value type: %v",
			v.Type().String())
	}
	p = &pb.Property{
		Name:     proto.String(name),
		Value:    &pv,
		Multiple: proto.Bool(multiple),
	}
	if v.IsValid() {
		switch v.Interface().(type) {
		case []byte:
			p.Meaning = pb.NewProperty_Meaning(pb.Property_BLOB)
		case appengine.BlobKey:
			p.Meaning = pb.NewProperty_Meaning(pb.Property_BLOBKEY)
		case time.Time:
			p.Meaning = pb.NewProperty_Meaning(pb.Property_GD_WHEN)
		}
	}
	return p, nil
}

// saveEntity saves an EntityProto into a PropertyLoadSaver or struct pointer.
func saveEntity(key *Key, src interface{}) (x *pb.EntityProto, err error) {
	c := make(chan Property, 32)
	donec := make(chan struct{})
	go func() {
		x, err = propertiesToProto(key, c)
		close(donec)
	}()
	var err1 error
	if e, ok := src.(PropertyLoadSaver); ok {
		err1 = e.Save(c)
	} else {
		err1 = SaveStruct(src, c)
	}
	<-donec
	if err1 != nil {
		return nil, err1
	}
	return x, err
}

func saveStructProperty(c chan<- Property, name string, noIndex, multiple bool, v reflect.Value) error {
	p := Property{
		Name:     name,
		NoIndex:  noIndex,
		Multiple: multiple,
	}
	switch x := v.Interface().(type) {
	case *Key:
		p.Value = x
	case time.Time:
		p.Value = x
	case appengine.BlobKey:
		p.Value = x
	case []byte:
		p.NoIndex = true
		p.Value = x
	default:
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			p.Value = v.Int()
		case reflect.Bool:
			p.Value = v.Bool()
		case reflect.String:
			p.Value = v.String()
		case reflect.Float32, reflect.Float64:
			p.Value = v.Float()
		}
	}
	if p.Value == nil {
		return fmt.Errorf("datastore: unsupported struct field type: %v",
			v.Type())
	}
	c <- p
	return nil
}

func (s structPLS) Save(c chan<- Property) error {
	defer close(c)
	for i, t := range s.codec.byIndex {
		if t.name == "-" {
			continue
		}
		v := s.v.Field(i)
		if !v.IsValid() || !v.CanSet() {
			continue
		}
		// For slice fields that aren't []byte, save each element.
		if v.Kind() == reflect.Slice && v.Type() != typeOfByteSlice {
			for j := 0; j < v.Len(); j++ {
				if err := saveStructProperty(c, t.name, t.noIndex, true, v.Index(j)); err != nil {
					return err
				}
			}
			continue
		}
		// Otherwise, save the field itself.
		if err := saveStructProperty(c, t.name, t.noIndex, false, v); err != nil {
			return err
		}
	}
	return nil
}

func propertiesToProto(key *Key, src <-chan Property) (*pb.EntityProto, error) {
	defer func() {
		for _ = range src {
			// Drain the src channel, if we exit early.
		}
	}()
	e := &pb.EntityProto{
		Key: keyToProto(key),
	}
	if key.parent == nil {
		e.EntityGroup = &pb.Path{}
	} else {
		e.EntityGroup = keyToProto(key.root()).Path
	}
	prevMultiple := make(map[string]bool)

	for p := range src {
		if pm, ok := prevMultiple[p.Name]; ok {
			if !pm || !p.Multiple {
				return nil, fmt.Errorf("datastore: multiple Properties with Name %q, but Multiple is false", p.Name)
			}
		} else {
			prevMultiple[p.Name] = p.Multiple
		}

		x := &pb.Property{
			Name:     proto.String(p.Name),
			Value:    new(pb.PropertyValue),
			Multiple: proto.Bool(p.Multiple),
		}
		switch v := p.Value.(type) {
		case int64:
			x.Value.Int64Value = proto.Int64(v)
		case bool:
			x.Value.BooleanValue = proto.Bool(v)
		case string:
			x.Value.StringValue = proto.String(v)
		case float64:
			x.Value.DoubleValue = proto.Float64(v)
		case *Key:
			if v != nil {
				x.Value.Referencevalue = keyToReferenceValue(v)
			}
		case time.Time:
			if v.Before(minTime) || v.After(maxTime) {
				return nil, errors.New("datastore: time value out of range")
			}
			x.Value.Int64Value = proto.Int64(v.UnixNano() / 1e3)
			x.Meaning = pb.NewProperty_Meaning(pb.Property_GD_WHEN)
		case appengine.BlobKey:
			x.Value.StringValue = proto.String(string(v))
			x.Meaning = pb.NewProperty_Meaning(pb.Property_BLOBKEY)
		case []byte:
			x.Value.StringValue = proto.String(string(v))
			x.Meaning = pb.NewProperty_Meaning(pb.Property_BLOB)
			if !p.NoIndex {
				return nil, fmt.Errorf("datastore: cannot index a []byte valued Property with Name %q", p.Name)
			}
		default:
			return nil, fmt.Errorf("datastore: invalid Value type for a Property with Name %q", p.Name)
		}

		if p.NoIndex {
			e.RawProperty = append(e.RawProperty, x)
		} else {
			e.Property = append(e.Property, x)
			if len(e.Property) > maxIndexedProperties {
				return nil, errors.New("datastore: too many indexed properties")
			}
		}
	}
	return e, nil
}
