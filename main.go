package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	schema "./schema"
	"github.com/gorilla/mux"
	graphql "github.com/graph-gophers/graphql-go"
)

type State struct {
	config *Config
	jwt    *JwtProvider
	db     *Database
}

func (self *State) withContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "state", self)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

var (
	wait time.Duration
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO")
}

func main() {
	config := loadConfig(".")

	state := &State{
		config: config,
		jwt:    &JwtProvider{config.JwtState},
		db:     newDB(config),
	}

	discordOauth := newOauth(state)
	graphQL := GraphQL{
		state:   state,
		schema:  graphql.MustParseSchema(schema.GetRootSchema(), &Query{}),
		loaders: newLoaderCollection(),
	}

	// Server setup
	log.Println("Starting server…")

	router := mux.NewRouter()
	server := &http.Server{
		Addr:         config.Address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	router.Use(loggingMiddleware)
	router.Use(corsMiddleware) // FIXME: Remove this
	router.Use(state.withContext())
	router.Use(state.jwt.middleware)

	fs := http.FileServer(http.Dir("../grip/build"))
	static := http.FileServer(http.Dir("../grip/build/static"))
	router.Handle("/", http.StripPrefix("/", fs))
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", static))

	router.HandleFunc("/graphiql", graphiqlHandler)
	jwtRoute := router.PathPrefix("/jwt").Subrouter()
	// jwtRoute.Use(state.jwt.middleware)

	// TODO: Delete this later
	jwtRoute.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if authenticated, _ := ctx.Value("authorized").(bool); authenticated {
			message := "Tested an authenticated user"

			fmt.Println(message)
			fmt.Fprintln(w, ctx)
		} else {
			message := "Tested user is not authenticated"

			fmt.Println(message)
			fmt.Fprintln(w, ctx)
		}
	})

	discordOauth.registerRoutes(router)
	graphQL.registerRoutes(router)

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Println("Server started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	<-c // Block until interrupt

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	server.Shutdown(ctx)

	log.Println("Shutting down…")
	os.Exit(0)
}
