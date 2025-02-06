// Hello world webserver based on https://www.jetbrains.com/guide/go/tutorials/rest_api_series/stdlib/

package main

import "net/http"

func main() {

	// Create a new request multiplexer
	// Take incoming requests and dispatch them to the matching handlers
	mux := http.NewServeMux()

	// register the routes and handlers
	mux.Handle("/", &homeHandler{})

	// Run the server
	http.ListenAndServe(":8080", mux)
}

type homeHandler struct{}

// define handler for homepage oute
func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my home page"))
}
