package datastore

import (
	"fmt"
	"gae-go-testing.googlecode.com/git/appenginetesting"
	"testing"
)

func getContext(t *testing.T) *appenginetesting.Context {
	c, err := appenginetesting.NewContext(nil)
	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	return c
}

func TestNamespaceKey(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns1")
	k2 := NewNamespaceKey(c, "Test", "foo", 0, k1, "")
	k3 := NewNamespaceKey(c, "Test", "foo", 0, nil, "")

	p1 := keyToProto(k1)
	k1, _ = protoToKey(p1)
	p2 := keyToProto(k2)
	k2, _ = protoToKey(p2)
	p3 := keyToProto(k3)
	k3, _ = protoToKey(p3)

	r1 := keyToReferenceValue(k1)
	k1, _ = referenceValueToKey(r1)
	r2 := keyToReferenceValue(k2)
	k2, _ = referenceValueToKey(r2)
	r3 := keyToReferenceValue(k3)
	k3, _ = referenceValueToKey(r3)

	if k1.Namespace() != "ns1" {
		t.Fatalf("Wrong namespace %v, expected %v", k1.Namespace(), "ns1")
	}
	if k2.Namespace() != "ns1" {
		t.Fatalf("Wrong namespace %v, expected %v", k2.Namespace(), "ns1")
	}
	if k3.Namespace() != "" {
		t.Fatalf("Wrong namespace %v, expected %v", k3.Namespace(), "")
	}
}

func TestNamespaceKeyEquality(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns1")
	k2 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns1")
	k3 := NewNamespaceKey(c, "Test", "foo", 0, k1, "")
	k4 := NewNamespaceKey(c, "Test", "foo", 0, k2, "")
	k5 := NewNamespaceKey(c, "Test", "foo", 0, nil, "")
	k6 := NewNamespaceKey(c, "Test", "foo", 0, nil, "")
	k7 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns2")

	if !k1.Equal(k2) {
		t.Fatalf("These keys are equal: %v, %v", k1, k2)
	}
	if !k3.Equal(k4) {
		t.Fatalf("These keys are equal: %v, %v", k3, k4)
	}
	if !k5.Equal(k6) {
		t.Fatalf("These keys are equal: %v, %v", k5, k6)
	}
	if k1.Equal(k3) {
		t.Fatalf("These keys are not equal: %v, %v", k1, k3)
	}
	if k1.Equal(k5) {
		t.Fatalf("These keys are not equal: %v, %v", k1, k5)
	}
	if k1.Equal(k7) {
		t.Fatalf("These keys are not equal: %v, %v", k1, k7)
	}
}

// ----------------------------------------------------------------------------

func getKeyMap(t *testing.T, iter *Iterator) map[string]*Key {
	m := make(map[string]*Key)
	for {
		key, err := iter.Next(&struct{}{})
		if err != nil {
			if err == Done {
				break
			}
			t.Errorf("Error on Run(): %v\n", err)
			break
		}
		m[key.Encode()] = key
	}
	return m
}

func TestKindlessQuery(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewKey(c, "A", "a", 0, nil)
	k2 := NewKey(c, "B", "b", 0, nil)
	k3 := NewKey(c, "C", "c", 0, nil)
	e := &struct{}{}
	if _, err := PutMulti(c, []*Key{k1, k2, k3}, []interface{}{e, e, e}); err != nil {
		t.Errorf("Error on PutMulti(): %v\n", err)
	}

	// Order on __key__ ascending.
	q1 := NewQuery("").Order("__key__").Limit(10)
	m1 := getKeyMap(t, q1.Run(c))
	if len(m1) != 3 || m1[k1.Encode()] == nil || m1[k2.Encode()] == nil || m1[k3.Encode()] == nil {
		t.Errorf("Expected 3 results, got %v\n", m1)
	}

	// Filter on __key__.
	q2 := q1.Filter("__key__>", k1).Limit(10)
	m2 := getKeyMap(t, q2.Run(c))
	if len(m2) != 2 || m2[k2.Encode()] == nil || m2[k3.Encode()] == nil {
		t.Errorf("Expected 2 results, got %v\n", m2)
	}
}

func TestKindlessAncestorQuery(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewKey(c, "A", "a", 0, nil)
	k2 := NewKey(c, "B", "b", 0, k1)
	k3 := NewKey(c, "C", "c", 0, k2)
	e := &struct{}{}
	if _, err := PutMulti(c, []*Key{k1, k2, k3}, []interface{}{e, e, e}); err != nil {
		t.Errorf("Error on PutMulti(): %v\n", err)
	}

	// Order on __key__ ascending.
	q1 := NewQuery("").Order("__key__").Ancestor(k1).Limit(10)
	m1 := getKeyMap(t, q1.Run(c))
	if len(m1) != 3 || m1[k1.Encode()] == nil || m1[k2.Encode()] == nil || m1[k3.Encode()] == nil {
		t.Errorf("Expected 3 results, got %v\n", m1)
	}

	// Filter on __key__.
	q2 := q1.Filter("__key__>", k1).Limit(10)
	m2 := getKeyMap(t, q2.Run(c))
	if len(m2) != 2 || m2[k2.Encode()] == nil || m2[k3.Encode()] == nil {
		t.Errorf("Expected 2 results, got %v\n", m2)
	}
}

// ----------------------------------------------------------------------------

func TestGetCursorAt(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	e := &struct{}{}
	keys := make([]*Key, 20)
	entities := make([]interface{}, 20)
	for i := 0; i < 20; i++ {
		keys[i] = NewKey(c, "A", fmt.Sprintf("%03d", i), 0, nil)
		entities[i] = e
	}
	if _, err := PutMulti(c, keys, entities); err != nil {
		t.Errorf("Error on PutMulti(): %v\n", err)
	}

	q1 := NewQuery("A").Limit(1).Compile(true)
	for i := 0; i < 20; i++ {
		cursor, err := q1.GetCursorAt(c, i)
		if err != nil {
			t.Fatalf("Failed to return cursor from CursorAt()")
		}
		q2 := NewQuery("A").Limit(1).Cursor(cursor)
		i2 := q2.Run(c)
		k2, err2 := i2.Next(struct{}{})
		if err2 == Done {
			break
		} else if err2 != nil {
			t.Fatalf("%v", err2)
		}
		if k2.StringID() != keys[i].StringID() {
			t.Errorf("Expected id %q, got %q", keys[i].StringID(), k2.StringID())
		}
	}
}

/*
func TestCursor(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	e := &struct{}{}
	keys := make([]*Key, 50)
	entities := make([]interface{}, 50)
	for i := 0; i < 50; i++ {
		keys[i] = NewKey(c, "A", fmt.Sprintf("%03d", i), 0, nil)
		entities[i] = e
	}

	if _, err := PutMulti(c, keys, entities); err != nil {
		t.Errorf("Error on PutMulti(): %v\n", err)
	}

	q1 := NewQuery("A").Limit(10).Cursor(nil)
	i1 := q1.Run(c)
	k1, _ := i1.Next(struct{}{})
	c1 := i1.GetCursorAfter()
	if k1.StringID() != "000" {
		t.Errorf("Expected %q string id, got %q", "000", k1.StringID())
	}

	q2 := NewQuery("A").Limit(10).Cursor(c1)
	i2 := q2.Run(c)
	k2, _ := i2.Next(struct{}{})
	c2 := i2.GetCursorAfter()
	if k2.StringID() != "010" {
		t.Errorf("Expected %q string id, got %q", "010", k2.StringID())
	}

	q3 := NewQuery("A").Limit(10).Cursor(c2)
	i3 := q3.Run(c)
	//c3 := i3.GetCursorAfter()
	k3, _ := i3.Next(struct{}{})
	if k3.StringID() != "020" {
		t.Errorf("Expected %q string id, got %q", "020", k3.StringID())
	}

	/*
	q4 := NewQuery("A").Limit(10).KeysOnly(true).Cursor(c3)
	k4, _, hasMore4, err4 := q4.GetPage(c, nil)
	if err4 != nil {
		t.Fatalf("Error on GetPage: %v", err4)
	}
	if len(k4) != 10 {
		t.Fatalf("Expected 10 results, got %v", len(k4))
	}

	if k4[0].StringID() != "030" {
		t.Errorf("Expected %q string id, got %q", "030", k4[0].StringID())
	}
	if k4[9].StringID() != "039" {
		t.Errorf("Expected %q string id, got %q", "039", k4[9].StringID())
	}
	if !hasMore4 {
		t.Errorf("Expected hasMore")
	}
}
*/
