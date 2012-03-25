// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"strings"

	"appengine"
)

// NewQuery creates a new Query for a specific entity kind.
func NewQuery(kind string) *Query {
	return &Query{base: NewBaseQuery().Kind(kind)}
}

// Query represents a datastore query.
type Query struct {
	base    *BaseQuery
	aliases map[string]string
}

// Clone returns a copy of the query.
func (q *Query) Clone() *Query {
	return &Query{base: q.base.Clone(), aliases: q.aliases}
}

// SetPropertyAliases sets a map of aliases for properties used in filters
// and orders.
func (q *Query) SetPropertyAliases(aliases map[string]string) *Query {
	q.aliases = aliases
	return q
}

// propertyName returns the name for a property given its alias.
func (q *Query) propertyName(alias string) string {
	if q.aliases != nil {
		if name, ok := q.aliases[alias]; ok {
			return name
		}
	}
	return alias
}

// Namespace sets the namespace for the query.
//
// This is a temporary function to fill the gap until appengine.Context
// supports namespaces.
func (q *Query) Namespace(namespace string) *Query {
	q.base.Namespace(namespace)
	return q
}

// Ancestor sets the ancestor filter for the query.
func (q *Query) Ancestor(key *Key) *Query {
	q.base.Ancestor(key)
	return q
}

// Kind sets the entity kind for the query.
func (q *Query) Kind(kind string) *Query {
	q.base.Kind(kind)
	return q
}

// Filter adds a field-based filter to the query.
// The filterStr argument must be a field name followed by optional space,
// followed by an operator, one of ">", "<", ">=", "<=", or "=".
// Fields are compared against the provided value using the operator.
// Multiple filters are AND'ed together.
func (q *Query) Filter(filter string, value interface{}) *Query {
	property := strings.TrimRight(filter, " ><=")
	var operator queryOperator
	switch strings.TrimSpace(filter[len(property):]) {
	case "<":
		operator = QueryOperatorLessThan
	case "<=":
		operator = QueryOperatorLessThanOrEqual
	case "=":
		operator = QueryOperatorEqual
	case ">=":
		operator = QueryOperatorGreaterThanOrEqual
	case ">":
		operator = QueryOperatorGreaterThan
	default:
		q.base.err = fmt.Errorf("datastore: invalid query filter %q",
			filter)
		return q
	}
	q.base.Filter(q.propertyName(property), operator, value)
	return q
}

// Order adds a field-based sort to the query.
// Orders are applied in the order they are added.
// The default order is ascending; to sort in descending
// order prefix the fieldName with a minus sign (-).
func (q *Query) Order(order string) *Query {
	property := order
	direction := QueryDirectionAscending
	if strings.HasPrefix(order, "-") {
		property = strings.TrimSpace(order[1:])
		direction = QueryDirectionDescending
	}
	q.base.Order(q.propertyName(property), direction)
	return q
}

// Limit sets the maximum number of keys/entities to return.
// A zero value means unlimited. A negative value is invalid.
func (q *Query) Limit(limit int) *Query {
	q.base.Limit(limit)
	return q
}

// Offset sets how many keys to skip over before returning results.
// A negative value is invalid.
func (q *Query) Offset(offset int) *Query {
	q.base.Offset(offset)
	return q
}

// KeysOnly configures the query to return keys, instead of keys and entities.
func (q *Query) KeysOnly(keysOnly bool) *Query {
	q.base.KeysOnly(keysOnly)
	return q
}

// Compile configures the query to produce cursors.
func (q *Query) Compile(compile bool) *Query {
	q.base.Compile(compile)
	return q
}

// Cursor sets the cursor position to start the query.
//
// When a cursor is set the query is automatically configured to compile.
func (q *Query) Cursor(cursor *Cursor) *Query {
	q.base.Cursor(cursor)
	return q
}

// Run runs the query in the given context.
func (q *Query) Run(c appengine.Context) *Iterator {
	return q.base.Run(c)
}

// GetAll runs the query in the given context and returns all keys that match
// that query, as well as appending the values to dst.
//
// dst must have type *[]S or *[]*S or *[]P, for some struct type S or some non-
// interface, non-pointer type P such that P or *P implements PropertyLoadSaver.
//
// As a special case, *PropertyList is an invalid type for dst, even though a
// PropertyList is a slice of structs. It is treated as invalid to avoid being
// mistakenly passed when *[]PropertyList was intended.
//
// If q is a ``keys-only'' query, GetAll ignores dst and only returns the keys.
func (q *Query) GetAll(c appengine.Context, dst interface{}) ([]*Key, error) {
	return q.base.GetAll(c, dst)
}

// GetPage is the same as GetAll, but it also returns a cursor and a flag
// indicating if there are more results.
func (q *Query) GetPage(c appengine.Context, dst interface{}) (keys []*Key,
	cursor *Cursor, hasMore bool, err error) {
	return q.base.GetPage(c, dst)
}

// Count returns the number of results for the query.
func (q *Query) Count(c appengine.Context) (int, error) {
	return q.base.Count(c)
}

// GetCursorAt returns a cursor at the given position for this query.
func (q *Query) GetCursorAt(c appengine.Context, position int) (*Cursor, error) {
	return q.base.GetCursorAt(c, position)
}
