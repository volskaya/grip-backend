package main

import (
	"context"
	"fmt"
	"log"

	"gopkg.in/nicksrandall/dataloader.v5"
)

type key string

const (
	userLoaderKey    key = "user"
	projectLoaderKey key = "project"
	raitingLoaderKey key = "raiting"
)

type LoaderCollection struct {
	dataloaderFuncMap map[key]dataloader.BatchFunc
}

func newLoaderCollection() LoaderCollection {
	userLoader := &UserLoader{}
	projectLoader := &ProjectLoader{}
	raitingLoader := &RaitingLoader{}

	return LoaderCollection{
		dataloaderFuncMap: map[key]dataloader.BatchFunc{
			userLoaderKey:    userLoader.loadBatch,
			projectLoaderKey: projectLoader.loadBatch,
			raitingLoaderKey: raitingLoader.loadBatch,
		},
	}
}

func (self *LoaderCollection) attach(ctx context.Context) context.Context {
	for key, batchFunc := range self.dataloaderFuncMap {
		ctx = context.WithValue(ctx, key, dataloader.NewBatchedLoader(batchFunc))
	}

	return ctx
}

func extract(ctx context.Context, k key) (*dataloader.Loader, error) {
	if ldr, ok := ctx.Value(k).(*dataloader.Loader); ok {
		return ldr, nil
	}

	return nil, fmt.Errorf("Unabled to extract %s loader from context", k)
}

func loadSomething(ctx context.Context, key string, loader key) (interface{}, error) {
	ldr, err := extract(ctx, loader)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	data := ldr.Load(ctx, dataloader.StringKey(key))
	res, err := data()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return res, nil
}

// TODO: Is this worth it?
// func mapItems(items interface{}) *map[string]interface{} {
// 	mapped := make(map[string]interface{})
// 	ref := reflect.ValueOf(items)

// 	for i := 0; i < ref.Len(); i++ {
// 		item := ref.Index(i)
// 		mapped[item.Elem().Field(0).String()] = item.Elem().Interface()
// 	}

// 	return &mapped
// }

// User loader
//------------------------------------------------------------------------------

type UserLoader struct{}

func (self *UserLoader) loadBatch(
	ctx context.Context,
	keys dataloader.Keys,
) []*dataloader.Result {
	var (
		n       = len(keys)
		results = make([]*dataloader.Result, n)
		ids     = make([]string, n)
		db      = ctx.Value("state").(*State).db
	)

	log.Printf("Fetching %s from UserLoader\n", keys)

	// Cheaper version, if only 1 key
	if n == 1 {
		var user User

		if _, err := db.findID(&user, keys[0].String()); err == nil {
			results[0] = &dataloader.Result{Data: user, Error: nil}
		} else {
			results[0] = &dataloader.Result{Data: nil, Error: nil}
		}

		return results
	}

	for i, key := range keys {
		ids[i] = key.String()
	}

	var users Users
	_, err := db.findWithID(&users, ids)

	if err != nil {
		for i := 0; i < n; i++ {
			results[i] = &dataloader.Result{Data: nil, Error: nil}
		}

		return results
	}

	mapped := make(map[string]*User)

	for _, user := range users {
		mapped[user.ID] = user
	}

	for i, id := range ids {
		if mapped[id] != nil {
			results[i] = &dataloader.Result{Data: *mapped[id], Error: nil}
		} else {
			results[i] = &dataloader.Result{Data: nil, Error: fmt.Errorf("user %s not found", id)}
		}
	}

	return results
}

// Project loader
//------------------------------------------------------------------------------

type ProjectLoader struct{}

func (self *ProjectLoader) loadBatch(
	ctx context.Context,
	keys dataloader.Keys,
) []*dataloader.Result {
	var (
		n       = len(keys)
		results = make([]*dataloader.Result, n)
		ids     = make([]string, n)
		db      = ctx.Value("state").(*State).db
	)

	log.Printf("Fetching %s from ProjectLoader\n", keys)

	// Cheaper version, if only 1 key
	if n == 1 {
		var project Project

		if _, err := db.findID(&project, keys[0].String()); err == nil {
			results[0] = &dataloader.Result{Data: project, Error: nil}
		} else {
			results[0] = &dataloader.Result{Data: nil, Error: nil}
		}

		return results
	}

	for i, key := range keys {
		ids[i] = key.String()
	}

	var items Projects
	_, err := db.findWithID(&items, ids)

	if err != nil {
		for i := 0; i < n; i++ {
			results[i] = &dataloader.Result{Data: nil, Error: nil}
		}

		return results
	}

	mapped := make(map[string]*Project)

	for _, item := range items {
		mapped[item.String()] = item
	}

	for i, id := range ids {
		if mapped[id] != nil {
			project := mapped[id]

			results[i] = &dataloader.Result{Data: *project, Error: nil}
		} else {
			results[i] = &dataloader.Result{Data: nil, Error: fmt.Errorf("project %s not found", id)}
		}
	}

	return results
}

// Raiting loader
//------------------------------------------------------------------------------

type RaitingLoader struct{}

func (self *RaitingLoader) loadBatch(
	ctx context.Context,
	keys dataloader.Keys,
) []*dataloader.Result {
	var (
		n       = len(keys)
		results = make([]*dataloader.Result, n)
		ids     = make([]string, n)
		db      = ctx.Value("state").(*State).db
	)

	log.Printf("Fetching %s from RaitingLoader\n", keys)

	// Cheaper version, if only 1 key
	if n == 1 {
		var raiting Raiting

		if _, err := db.findID(&raiting, keys[0].String()); err == nil {
			results[0] = &dataloader.Result{Data: raiting, Error: nil}
		} else {
			results[0] = &dataloader.Result{Data: nil, Error: nil}
		}

		return results
	}

	for i, key := range keys {
		ids[i] = key.String()
	}

	var items Raitings
	_, err := db.findWithID(&items, ids)

	if err != nil {
		for i := 0; i < n; i++ {
			results[i] = &dataloader.Result{Data: nil, Error: nil}
		}

		return results
	}

	mapped := make(map[string]*Raiting)

	for _, item := range items {
		mapped[item.String()] = item
	}

	for i, id := range ids {
		if mapped[id] != nil {
			raiting := mapped[id]

			results[i] = &dataloader.Result{Data: *raiting, Error: nil}
		} else {
			results[i] = &dataloader.Result{Data: nil, Error: fmt.Errorf("project %s not found", id)}
		}
	}

	return results
}
