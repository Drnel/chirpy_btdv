package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {
	var serve_mux = http.NewServeMux()
	var server = http.Server{Handler: serve_mux}
	server.Addr = ":8080"
	var apiCfg = apiConfig{}
	serve_mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	serve_mux.Handle("/metrics", apiCfg.middlewareMetricsShow())
	serve_mux.Handle("/reset", apiCfg.middlewareMetricsReset())
	serve_mux.HandleFunc("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}))

	fmt.Println("Starting Chirpy server:")
	server.ListenAndServe()
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsShow() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
	})
}

func (cfg *apiConfig) middlewareMetricsReset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
	})
}
