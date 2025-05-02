package main

import (
	"fmt"
	"net/http"
)

func main() {
	var serve_mux = http.NewServeMux()
	var server = http.Server{Handler: serve_mux}
	server.Addr = ":8080"
	serve_mux.Handle("/", http.FileServer(http.Dir(".")))
	fmt.Println("Starting Chirpy server:")
	server.ListenAndServe()
}
