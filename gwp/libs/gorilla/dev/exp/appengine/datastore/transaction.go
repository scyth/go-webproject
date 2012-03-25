// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"errors"
	"reflect"

	"appengine"
	"appengine_internal"
	"code.google.com/p/goprotobuf/proto"

	pb "appengine_internal/datastore"
)

// ErrConcurrentTransaction is returned when a transaction is rolled back due
// to a conflict with a concurrent transaction.
var ErrConcurrentTransaction = errors.New("datastore: concurrent transaction")

type transaction struct {
	appengine.Context
	transaction pb.Transaction
	finished    bool
}

var errBadTransactionField = errors.New("datastore: Call parameter has an incompatible Transaction field")

// setTransactionField performs the equivalent of "x.Transaction =
// &t.transaction", where x is a pointer to a protocol buffer value whose
// Transaction field has a different but field-for-field equivalent protocol
// buffer type to the *pb.Transaction type.
//
// It is not an error if x does not have a Transaction field at all.
//
// errBadTransactionField is returned if x does have a Transaction field but
// it is not field-for-field assignment compatible with t.transaction.
//
// This reflect-based copy is necessary because the script that generates the
// datastore and taskqueue protocol buffer definitions lead to two distinct
// types in Go, even though they are conceptually the same.
func (t *transaction) setTransactionField(x interface{}) (err error) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr || v.Type().Elem().Kind() != reflect.Struct {
		return errBadTransactionField
	}
	v = v.Elem().FieldByName("Transaction")
	if !v.IsValid() {
		// x does not have a Transaction field. That is not an error.
		return nil
	}
	if v.Kind() != reflect.Ptr || v.Type().Elem().Kind() != reflect.Struct {
		return errBadTransactionField
	}
	// Perform a field-for-field copy from t.transaction to a newly allocated value.
	fieldPtr := v
	dstPtr := reflect.New(v.Type().Elem())
	dst := dstPtr.Elem()
	src := reflect.ValueOf(&t.transaction).Elem()
	if dst.NumField() != src.NumField() {
		return errBadTransactionField
	}
	for i := 0; i < src.NumField(); i++ {
		if dst.Type().Field(i).Name != src.Type().Field(i).Name {
			return errBadTransactionField
		}
		df := dst.Field(i)
		sf := src.Field(i)
		if !sf.Type().AssignableTo(df.Type()) {
			return errBadTransactionField
		}
		df.Set(sf)
	}
	fieldPtr.Set(dstPtr)
	return nil
}

func (t *transaction) Call(service, method string, in, out interface{}, opts *appengine_internal.CallOptions) error {
	if t.finished {
		return errors.New("datastore: transaction context has expired")
	}
	switch service {
	case "datastore_v3":
		switch x := in.(type) {
		case *pb.Query:
			x.Transaction = &t.transaction
		case *pb.GetRequest:
			x.Transaction = &t.transaction
		case *pb.PutRequest:
			x.Transaction = &t.transaction
		case *pb.DeleteRequest:
			x.Transaction = &t.transaction
		}
	case "taskqueue":
		if err := t.setTransactionField(in); err != nil {
			return err
		}
	}
	return t.Context.Call(service, method, in, out, opts)
}

func runOnce(c appengine.Context, f func(appengine.Context) error, opts *TransactionOptions) error {
	// Begin the transaction.
	t := &transaction{Context: c}
	req := &pb.BeginTransactionRequest{
		App: proto.String(c.FullyQualifiedAppID()),
	}
	if opts != nil && opts.XG {
		req.AllowMultipleEg = proto.Bool(true)
	}
	if err := t.Context.Call("datastore_v3", "BeginTransaction", req, &t.transaction, nil); err != nil {
		return err
	}

	// Call f, rolling back the transaction if f returns a non-nil error, or panics.
	// The panic is not recovered.
	defer func() {
		if t.finished {
			return
		}
		t.finished = true
		// Ignore the error return value, since we are already returning a non-nil
		// error (or we're panicking).
		c.Call("datastore_v3", "Rollback", &t.transaction, &pb.VoidProto{}, nil)
	}()
	if err := f(t); err != nil {
		return err
	}
	t.finished = true

	// Commit the transaction.
	res := &pb.CommitResponse{}
	err := c.Call("datastore_v3", "Commit", &t.transaction, res, nil)
	if ae, ok := err.(*appengine_internal.APIError); ok {
		if appengine.IsDevAppServer() {
			// The Python Dev AppServer raises an ApplicationError with error code 2 (which is
			// Error.CONCURRENT_TRANSACTION) and message "Concurrency exception.".
			if ae.Code == int32(pb.Error_BAD_REQUEST) && ae.Detail == "ApplicationError: 2 Concurrency exception." {
				return ErrConcurrentTransaction
			}
		}
		if ae.Code == int32(pb.Error_CONCURRENT_TRANSACTION) {
			return ErrConcurrentTransaction
		}
	}
	return err
}

// RunInTransaction runs f in a transaction. It calls f with a transaction
// context tc that f should use for all App Engine operations.
//
// If f returns nil, RunInTransaction attempts to commit the transaction,
// returning nil if it succeeds. If the commit fails due to a conflicting
// transaction, RunInTransaction retries f, each time with a new transaction
// context. It gives up and returns ErrConcurrentTransaction after three
// failed attempts.
//
// If f returns non-nil, then any datastore changes will not be applied and
// RunInTransaction returns that same error. The function f is not retried.
//
// Note that when f returns, the transaction is not yet committed. Calling code
// must be careful not to assume that any of f's changes have been committed
// until RunInTransaction returns nil.
//
// Nested transactions are not supported; c may not be a transaction context.
func RunInTransaction(c appengine.Context, f func(tc appengine.Context) error, opts *TransactionOptions) error {
	if _, ok := c.(*transaction); ok {
		return errors.New("datastore: nested transactions are not supported")
	}
	for i := 0; i < 3; i++ {
		if err := runOnce(c, f, opts); err != ErrConcurrentTransaction {
			return err
		}
	}
	return ErrConcurrentTransaction
}

// TransactionOptions are the options for running a transaction.
type TransactionOptions struct {
	// XG is whether the transaction can cross multiple entity groups. In
	// comparison, a single group transaction is one where all datastore keys
	// used have the same root key. Note that cross group transactions do not
	// have the same behavior as single group transactions. In particular, it
	// is much more likely to see partially applied transactions in different
	// entity groups, in global queries.
	// It is valid to set XG to true even if the transaction is within a
	// single entity group.
	XG bool
}
