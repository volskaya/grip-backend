package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

// Extending ResponseWriter, to store responses status code
type StatusWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *StatusWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := new(bytes.Buffer)
		sw := &StatusWriter{w, http.StatusOK}

		fmt.Fprintf(
			b,
			"%s \"%s %s %s\" %d %s %d \"%s\" \"%s\"",
			r.RemoteAddr,
			r.Method,
			r.URL,
			r.Proto,
			sw.StatusCode,
			http.StatusText(sw.StatusCode),
			r.ContentLength,
			r.RequestURI,
			r.Header.Get("User-Agent"),
		)

		log.Println(b)
		next.ServeHTTP(sw, r)
	})
}

// FIXME: Remove this
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, Accept-Encoding")

		if r.Method == "OPTIONS" {
			fmt.Println("Giving cors")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
