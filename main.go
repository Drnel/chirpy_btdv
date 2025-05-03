package main

import (
	"fmt"
	"net/http"
)

func main() {
	var serve_mux = http.NewServeMux()
	var server = http.Server{Handler: serve_mux}
	server.Addr = ":8080"

	serve_mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	serve_mux.HandleFunc("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}))

	fmt.Println("Starting Chirpy server:")
	server.ListenAndServe()
}
