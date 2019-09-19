package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	graphql "github.com/graph-gophers/graphql-go"
)

type GraphQL struct {
	state   *State
	schema  *graphql.Schema
	loaders LoaderCollection
}

func (self *GraphQL) serve(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := self.loaders.attach(r.Context())
	response := self.schema.Exec(
		ctx,
		params.Query,
		params.OperationName,
		params.Variables,
	)

	responseJSON, err := json.Marshal(response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func (self *GraphQL) registerRoutes(router *mux.Router) {
	router.HandleFunc("/graphql", self.serve)
}
