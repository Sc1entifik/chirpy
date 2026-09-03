package main

import (
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
