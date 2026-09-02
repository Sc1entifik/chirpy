package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) returnRequests() int32 {

	return cfg.fileserverHits.Load()
}

func (cfg *apiConfig) resetRequests() {
	cfg.fileserverHits.Store(0)
}


func main() {
	apiCfg := &apiConfig{}
	mux := http.NewServeMux()
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	
	mux.Handle("GET /app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /api/metrics", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("Hits: %d", apiCfg.returnRequests())))
	}) 
	
	mux.HandleFunc("POST /api/reset", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		apiCfg.resetRequests()
		w.Write([]byte(fmt.Sprintf("Server Hits reset! %d", apiCfg.returnRequests())))
	})

	server.ListenAndServe()
}
