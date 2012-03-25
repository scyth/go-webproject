// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	"code.google.com/p/goprotobuf/proto"

	"appengine"
	pb "appengine_internal/datastore"
)

type queryOperator int

// Filter operators.
const (
	QueryOperatorLessThan queryOperator = iota
	QueryOperatorLessThanOrEqual
	QueryOperatorEqual
	QueryOperatorGreaterThanOrEqual
	QueryOperatorGreaterThan
)

var queryOperatorToProto = map[queryOperator]*pb.Query_Filter_Operator{
	QueryOperatorLessThan:           pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN),
	QueryOperatorLessThanOrEqual:    pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN_OR_EQUAL),
	QueryOperatorEqual:              pb.NewQuery_Filter_Operator(pb.Query_Filter_EQUAL),
	QueryOperatorGreaterThanOrEqual: pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN_OR_EQUAL),
	QueryOperatorGreaterThan:        pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN),
}

type queryDirection int

// Order directions.
const (
	QueryDirectionAscending queryDirection = iota
	QueryDirectionDescending
)

var queryDirectionToProto = map[queryDirection]*pb.Query_Order_Direction{
	QueryDirectionAscending:  pb.NewQuery_Order_Direction(pb.Query_Order_ASCENDING),
	QueryDirectionDescending: pb.NewQuery_Order_Direction(pb.Query_Order_DESCENDING),
}

// ----------------------------------------------------------------------------
// BaseQuery
// ----------------------------------------------------------------------------

// NewBaseQuery returns a new BaseQuery.
func NewBaseQuery() *BaseQuery {
	return &BaseQuery{pbq: new(pb.Query)}
}

// BaseQuery deals with protocol buffers so that Query doesn't have to.
//
// It is a more bare version of datastore.Query intended to be used as base
// for query objects. It stores the query protobuf directly instead of
// building it when the query runs, and has a slightly less friendly API for
// filters and orders because it doesn't perform any string parsing --
// syntax flavors are left for implementations such as Query.
//
// When an error occurs, further method calls don't perform any operation.
type BaseQuery struct {
	pbq *pb.Query
	err error
}

// Clone returns a copy of the query.
func (q *BaseQuery) Clone() *BaseQuery {
	return &BaseQuery{pbq: &(*q.pbq), err: q.err}
}

// Namespace sets the namespace for the query.
func (q *BaseQuery) Namespace(namespace string) *BaseQuery {
	if q.err == nil {
		if namespace == "" {
			q.pbq.NameSpace = nil
		} else {
			q.pbq.NameSpace = proto.String(namespace)
		}
	}
	return q
}

// Ancestor sets the ancestor filter for the query.
func (q *BaseQuery) Ancestor(key *Key) *BaseQuery {
	if q.err == nil {
		if key == nil {
			q.pbq.Ancestor = nil
		} else {
			if key.Incomplete() {
				q.err = errors.New("datastore: incomplete query ancestor key")
			} else {
				q.pbq.Ancestor = keyToProto(key)
			}
		}
	}
	return q
}

// Kind sets the entity kind for the query.
func (q *BaseQuery) Kind(kind string) *BaseQuery {
	if q.err == nil {
		if kind == "" {
			q.pbq.Kind = nil
		} else {
			q.pbq.Kind = proto.String(kind)
		}
	}
	return q
}

// Filter adds a field-based filter to the query.
func (q *BaseQuery) Filter(property string, operator queryOperator,
	value interface{}) *BaseQuery {
	if q.err == nil {
		var p *pb.Property
		p, q.err = valueToProto(property, value, false)
		if q.err == nil {
			q.pbq.Filter = append(q.pbq.Filter, &pb.Query_Filter{
				Op:       queryOperatorToProto[operator],
				Property: []*pb.Property{p},
			})
		}
	}
	return q
}

// Order adds a field-based sort to the query.
func (q *BaseQuery) Order(property string, direction queryDirection) *BaseQuery {
	if q.err == nil {
		q.pbq.Order = append(q.pbq.Order, &pb.Query_Order{
			Property:  proto.String(property),
			Direction: queryDirectionToProto[direction],
		})
	}
	return q
}

// Limit sets the maximum number of keys/entities to return.
// A zero value means unlimited. A negative value is invalid.
func (q *BaseQuery) Limit(limit int) *BaseQuery {
	if q.err == nil {
		if q.err = validateInt32(limit, "limit"); q.err == nil {
			q.pbq.Limit = proto.Int32(int32(limit))
		}
	}
	return q
}

// Offset sets how many keys to skip over before returning results.
// A negative value is invalid.
func (q *BaseQuery) Offset(offset int) *BaseQuery {
	if q.err == nil {
		if q.err = validateInt32(offset, "offset"); q.err == nil {
			q.pbq.Offset = proto.Int32(int32(offset))
		}
	}
	return q
}

// KeysOnly configures the query to return keys, instead of keys and entities.
func (q *BaseQuery) KeysOnly(keysOnly bool) *BaseQuery {
	if q.err == nil {
		q.pbq.KeysOnly = proto.Bool(keysOnly)
		q.pbq.RequirePerfectPlan = proto.Bool(keysOnly)
	}
	return q
}

// Compile configures the query to produce cursors.
func (q *BaseQuery) Compile(compile bool) *BaseQuery {
	if q.err == nil {
		q.pbq.Compile = proto.Bool(compile)
	}
	return q
}

// Cursor sets the cursor position to start the query.
func (q *BaseQuery) Cursor(cursor *Cursor) *BaseQuery {
	if q.err == nil {
		if cursor == nil {
			q.pbq.CompiledCursor = nil
		} else if cursor.compiled == nil {
			q.err = errors.New("datastore: empty start cursor")
		} else {
			q.pbq.CompiledCursor = cursor.compiled
		}
	}
	if q.err == nil {
		q.pbq.Compile = proto.Bool(true)
	}
	return q
}

// EndCursor sets the cursor position to end the query.
func (q *BaseQuery) EndCursor(cursor *Cursor) *BaseQuery {
	if q.err == nil {
		if cursor == nil {
			q.pbq.EndCompiledCursor = nil
		} else if cursor.compiled == nil {
			q.err = errors.New("datastore: empty end cursor")
		} else {
			q.pbq.EndCompiledCursor = cursor.compiled
		}
	}
	if q.err == nil {
		q.pbq.Compile = proto.Bool(true)
	}
	return q
}

// toProto converts the query to a protocol buffer.
//
// The zeroLimitMeansZero flag defines how to interpret a zero query/cursor
// limit. In some contexts, it means an unlimited query (to follow Go's idiom
// of a zero value being a useful default value). In other contexts, it means
// a literal zero, such as when issuing a query count, no actual entity data
// is wanted, only the number of skipped results.
func (q *BaseQuery) toProto(pbq *pb.Query, zeroLimitMeansZero bool) error {
	if q.err != nil {
		return q.err
	}
	if !zeroLimitMeansZero && proto.GetInt32(pbq.Limit) == 0 {
		pbq.Limit = nil
	}
	return nil
}

// Run runs the query in the given context.
func (q *BaseQuery) Run(c appengine.Context) *Iterator {
	// Make a copy of the query.
	req := *q.pbq
	if err := q.toProto(&req, false); err != nil {
		return &Iterator{err: q.err}
	}
	req.App = proto.String(c.FullyQualifiedAppID())
	t := &Iterator{
		c:      c,
		q:      q,
		limit:  proto.GetInt32(req.Limit),
		offset: proto.GetInt32(req.Offset),
	}
	if err := c.Call("datastore_v3", "RunQuery", &req, &t.res, nil); err != nil {
		t.err = err
		return t
	}
	return t
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
func (q *BaseQuery) GetAll(c appengine.Context, dst interface{}) ([]*Key, error) {
	var (
		dv       reflect.Value
		mat      multiArgType
		elemType reflect.Type
	)
	keysOnly := q.pbq.KeysOnly != nil && *q.pbq.KeysOnly
	if !keysOnly {
		dv = reflect.ValueOf(dst)
		if dv.Kind() != reflect.Ptr || dv.IsNil() {
			return nil, ErrInvalidEntityType
		}
		dv = dv.Elem()
		mat, elemType = checkMultiArg(dv)
		if mat == multiArgTypeInvalid || mat == multiArgTypeInterface {
			return nil, ErrInvalidEntityType
		}
	}

	var keys []*Key
	for t := q.Run(c); ; {
		k, e, err := t.next()
		if err == Done {
			break
		}
		if err != nil {
			return keys, err
		}
		if !keysOnly {
			ev := reflect.New(elemType)
			if elemType.Kind() == reflect.Map {
				// This is a special case. The zero values of a map type are
				// not immediately useful; they have to be make'd.
				//
				// Funcs and channels are similar, in that a zero value is not useful,
				// but even a freshly make'd channel isn't useful: there's no fixed
				// channel buffer size that is always going to be large enough, and
				// there's no goroutine to drain the other end. Theoretically, these
				// types could be supported, for example by sniffing for a constructor
				// method or requiring prior registration, but for now it's not a
				// frequent enough concern to be worth it. Programmers can work around
				// it by explicitly using Iterator.Next instead of the Query.GetAll
				// convenience method.
				x := reflect.MakeMap(elemType)
				ev.Elem().Set(x)
			}
			if err = loadEntity(ev.Interface(), e); err != nil {
				return keys, err
			}
			if mat != multiArgTypeStructPtr {
				ev = ev.Elem()
			}
			dv.Set(reflect.Append(dv, ev))
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// GetPage is the same as GetAll, but it also returns a cursor and a flag
// indicating if there are more results.
func (q *BaseQuery) GetPage(c appengine.Context, dst interface{}) (keys []*Key,
	cursor *Cursor, hasMore bool, err error) {
	q = q.Clone()
	limit := int(proto.GetInt32(q.pbq.Limit))
	q.Limit(limit + 1)
	if keys, err = q.GetAll(c, dst); err != nil {
		return nil, nil, false, err
	}
	if len(keys) > limit {
		hasMore = true
		keys = keys[:limit]
	}
	if cursor, err = q.GetCursorAt(c, limit); err != nil {
		return nil, nil, false, err
	}
	return
}

// Count returns the number of results for the query.
func (q *BaseQuery) Count(c appengine.Context) (int, error) {
	// Check that the query is well-formed.
	if q.err != nil {
		return 0, q.err
	}
	// Run a copy of the query, with keysOnly true, and an adjusted offset.
	// We also set the limit to zero, as we don't want any actual entity data,
	// just the number of skipped results.
	newQ := q.Clone()
	newQ.KeysOnly(true)
	newQ.Limit(0)
	limit := proto.GetInt32(q.pbq.Limit)
	offset := proto.GetInt32(q.pbq.Offset)
	var newOffset int32
	if limit == 0 {
		// If the original query was unlimited, set the new query's offset to maximum.
		newOffset = math.MaxInt32
	} else {
		newOffset = offset + limit
		if newOffset < 0 {
			// Do the best we can, in the presence of overflow.
			newOffset = math.MaxInt32
		}
	}
	newQ.Offset(int(newOffset))
	req := &pb.Query{}
	if err := newQ.toProto(req, true); err != nil {
		return 0, err
	}
	res := &pb.QueryResult{}
	if err := c.Call("datastore_v3", "RunQuery", req, res, nil); err != nil {
		return 0, err
	}

	// n is the count we will return. For example, suppose that our original
	// query had an offset of 4 and a limit of 2008: the count will be 2008,
	// provided that there are at least 2012 matching entities. However, the
	// RPCs will only skip 1000 results at a time. The RPC sequence is:
	//   call RunQuery with (offset, limit) = (2012, 0)  // 2012 == newQ.offset
	//   response has (skippedResults, moreResults) = (1000, true)
	//   n += 1000  // n == 1000
	//   call Next     with (offset, limit) = (1012, 0)  // 1012 == newQ.offset - n
	//   response has (skippedResults, moreResults) = (1000, true)
	//   n += 1000  // n == 2000
	//   call Next     with (offset, limit) = (12, 0)    // 12 == newQ.offset - n
	//   response has (skippedResults, moreResults) = (12, false)
	//   n += 12    // n == 2012
	//   // exit the loop
	//   n -= 4     // n == 2008
	var n int32
	for {
		// The QueryResult should have no actual entity data, just skipped results.
		if len(res.Result) != 0 {
			return 0, errors.New("datastore: internal error: Count request returned too much data")
		}
		n += proto.GetInt32(res.SkippedResults)
		if !proto.GetBool(res.MoreResults) {
			break
		}
		if err := callNext(c, res, newOffset-n, 0, true); err != nil {
			return 0, err
		}
	}
	n -= offset
	if n < 0 {
		// If the offset was greater than the number of matching entities,
		// return 0 instead of negative.
		n = 0
	}
	return int(n), nil
}

// GetCursorAt returns a cursor at the given position for this query.
func (q *BaseQuery) GetCursorAt(c appengine.Context, position int) (*Cursor, error) {
	if q.err != nil {
		return nil, q.err
	}
	if err := validateInt32(position, "cursor position"); err != nil {
		return nil, err
	}
	q = q.Clone().Limit(0).Offset(position).KeysOnly(true).Compile(true)
	t := q.Run(c)
	for {
		if _, err := t.Next(nil); err == Done {
			break
		} else if err != nil {
			return nil, err
		}
	}
	return t.getCursor(), nil
}

// validateInt32 validates that an int is positive ad doesn't overflow.
func validateInt32(v int, name string) error {
	if v < 0 {
		return fmt.Errorf("datastore: negative value for %v", name)
	}
	if v > math.MaxInt32 {
		return fmt.Errorf("datastore: value overflow for %v", name)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// Done is returned when a query iteration has completed.
var Done = errors.New("datastore: query has no more results")

// Iterator is the result of running a query.
type Iterator struct {
	c      appengine.Context
	q      *BaseQuery
	offset int32
	limit  int32
	res    pb.QueryResult
	curr   int // position of the current item in the current batch
	last   int // position of the last item in the current batch
	err    error
}

// Next returns the key of the next result. When there are no more results,
// Done is returned as the error.
// If the query is not keys only, it also loads the entity stored for that key
// into the struct pointer or PropertyLoadSaver dst, with the same semantics
// and possible errors as for the Get function.
// If the query is keys only, it is valid to pass a nil interface{} for dst.
func (t *Iterator) Next(dst interface{}) (*Key, error) {
	k, e, err := t.next()
	if err != nil || e == nil {
		return k, err
	}
	t.curr += 1
	return k, loadEntity(dst, e)
}

func (t *Iterator) next() (*Key, *pb.EntityProto, error) {
	if err := t.nextBatch(); err != nil {
		return nil, nil, err
	}
	// Pop the EntityProto from the front of t.res.Result and
	// extract its key.
	var e *pb.EntityProto
	e, t.res.Result = t.res.Result[0], t.res.Result[1:]
	if e.Key == nil {
		return nil, nil, errors.New("datastore: internal error: server did not return a key")
	}
	k, err := protoToKey(e.Key)
	if err != nil || k.Incomplete() {
		return nil, nil, errors.New("datastore: internal error: server returned an invalid key")
	}
	if proto.GetBool(t.res.KeysOnly) {
		return k, nil, nil
	}
	return k, e, nil
}

func (t *Iterator) nextBatch() error {
	if t.err != nil {
		return t.err
	}
	// Issue datastore_v3/Next RPCs as necessary.
	for len(t.res.Result) == 0 {
		if !proto.GetBool(t.res.MoreResults) {
			t.err = Done
			return t.err
		}
		t.offset -= proto.GetInt32(t.res.SkippedResults)
		if t.offset < 0 {
			t.offset = 0
		}
		if err := callNext(t.c, &t.res, t.offset, t.limit, false); err != nil {
			t.err = err
			return t.err
		}
		// Update last position counter.
		t.last += len(t.res.Result)
		// For an Iterator, a zero limit means unlimited.
		if t.limit == 0 {
			continue
		}
		t.limit -= int32(len(t.res.Result))
		if t.limit > 0 {
			continue
		}
		t.limit = 0
		if proto.GetBool(t.res.MoreResults) {
			t.err = errors.New("datastore: internal error: limit exhausted but more_results is true")
			return t.err
		}
	}
	return nil
}

// GetCursorAfter returns a cursor positioned just after the item returned by
// Iterator.Next().
//
// Note that sometimes requesting a cursor requires a datastore roundtrip
// (but not if the cursor corresponds to a batch boundary, normally
// the query limit).
//
// If Compile(true) was not set or a cursor wasn't set in the query, it
// always returns nil. If Next() wasn't called yet it also returns nil.
func (t *Iterator) GetCursorAfter() *Cursor {
	return t.getCursorAt(t.curr)
}

// CursorBefore returns a cursor positioned just before the item returned by
// Iterator.Next().
//
// Note that sometimes requesting a cursor requires a datastore roundtrip
// (but not if the cursor corresponds to a batch boundary, normally
// the query limit).
//
// If Compile(true) was not set or a cursor wasn't set in the query, it
// always returns nil. If Next() wasn't called yet it also returns nil.
func (t *Iterator) GetCursorBefore() *Cursor {
	return t.getCursorAt(t.curr - 1)
}

// getCursorAt returns a cursor in the given position.
func (t *Iterator) getCursorAt(position int) *Cursor {
	if err := t.nextBatch(); err != nil && err != Done {
		return nil
	}
	if t.curr == 0 || t.res.CompiledCursor == nil {
		// Next() wasn't called or query is not configured to compile.
		return nil
	}
	if position == t.last {
		// Cursor from the current batch.
		return t.getCursor()
	}
	// Perform datastore roundtrip.
	cursor, err := t.q.GetCursorAt(t.c, position)
	if err != nil {
		return nil
	}
	return cursor
}

// getCursor returns the cursor for the current batch.
func (t *Iterator) getCursor() *Cursor {
	return &Cursor{t.res.CompiledCursor}
}

// callNext issues a datastore_v3/Next RPC to advance a cursor, such as that
// returned by a query with more results.
func callNext(c appengine.Context, res *pb.QueryResult, offset, limit int32, zeroLimitMeansZero bool) error {
	if res.Cursor == nil {
		return errors.New("datastore: internal error: server did not return a cursor")
	}
	// TODO: should I eventually call datastore_v3/DeleteCursor on the cursor?
	req := &pb.NextRequest{
		Cursor: res.Cursor,
		Offset: proto.Int32(offset),
	}
	if limit != 0 || zeroLimitMeansZero {
		req.Count = proto.Int32(limit)
	}
	if res.CompiledCursor != nil {
		req.Compile = proto.Bool(true)
	}
	res.Reset()
	return c.Call("datastore_v3", "Next", req, res, nil)
}

// ----------------------------------------------------------------------------
// Cursor
// ----------------------------------------------------------------------------

// Cursor represents a compiled query cursor.
type Cursor struct {
	compiled *pb.CompiledCursor
}

// Encode returns an opaque representation of the cursor suitable for use in
// HTML and URLs. This is compatible with the Python and Java runtimes.
func (c *Cursor) Encode() string {
	if c.compiled != nil {
		if b, err := proto.Marshal(c.compiled); err == nil {
			// Trailing padding is stripped.
			return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
		}
	}
	// We don't return the error to follow Key.Encode which only
	// returns a string. It is unlikely to happen anyway.
	return ""
}

// DecodeCursor decodes a cursor from the opaque representation returned by
// Cursor.Encode.
func DecodeCursor(encoded string) (*Cursor, error) {
	// Re-add padding.
	if m := len(encoded) % 4; m != 0 {
		encoded += strings.Repeat("=", 4-m)
	}
	b, err := base64.URLEncoding.DecodeString(encoded)
	if err == nil {
		var c pb.CompiledCursor
		err = proto.Unmarshal(b, &c)
		if err == nil {
			return &Cursor{&c}, nil
		}
	}
	return nil, err
}
