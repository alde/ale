package mock

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"github.com/alde/ale"
)

type Datastore struct {
	memory map[string]*ale.JenkinsData

	PutFn        func(context.Context, *datastore.Key, interface{}) (*datastore.Key, error)
	PutFnInvoked bool

	GetFn        func(context.Context, *datastore.Key, interface{}) error
	GetFnInvoked bool

	CountFn        func(context.Context, *datastore.Query) (int, error)
	CountFnInvoked bool
}

// Put inserts data into the database
func (md *Datastore) defaultPutFn(ctx context.Context, key *datastore.Key, data interface{}) (*datastore.Key, error) {
	if md.memory == nil {
		md.memory = make(map[string]*ale.JenkinsData)
	}
	d := data.(*ale.DatastoreEntity)
	md.memory[d.Key] = &d.Value
	return nil, nil
}

// Get retrieves data from the database
func (md *Datastore) defaultGetFn(ctx context.Context, key *datastore.Key, data interface{}) error {
	if md.memory == nil {
		md.memory = make(map[string]*ale.JenkinsData)
	}
	data, ok := md.memory[key.Name]
	if !ok {
		return errors.New("not found")
	}
	return nil
}

func (md *Datastore) defaultCountFn(context.Context, *datastore.Query) (int, error) {
	if md.memory == nil {
		md.memory = make(map[string]*ale.JenkinsData)
		return 0, errors.New("not found")
	}
	return 1, nil
}

func (md *Datastore) Put(ctx context.Context, key *datastore.Key, data interface{}) (*datastore.Key, error) {
	md.PutFnInvoked = true
	if md.PutFn == nil {
		return md.defaultPutFn(ctx, key, data)
	}
	return md.PutFn(ctx, key, data)
}

func (md *Datastore) Get(ctx context.Context, key *datastore.Key, data interface{}) error {
	md.GetFnInvoked = true
	if md.GetFn == nil {
		return md.defaultGetFn(ctx, key, data)
	}
	return md.GetFn(ctx, key, data)
}

func (md *Datastore) Count(ctx context.Context, query *datastore.Query) (int, error) {
	md.CountFnInvoked = true
	if md.CountFn == nil {
		return md.defaultCountFn(ctx, query)
	}
	return md.CountFn(ctx, query)
}
