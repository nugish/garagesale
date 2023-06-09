package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	// Convert the Echo function to a type that implements http.Handler
	h := http.HandlerFunc(Echo)

	// Start a server listening on port 8000 and handle all requests with Echo
	log.Println("listeing on localhost:8000")
	if err := http.ListenAndServe("localhost:8000", h); err != nil {
		log.Fatal(err)
	}
}

// Echo is a basic HTTP handler
func Echo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "You asked to", r.Method, r.URL.Path)
}
