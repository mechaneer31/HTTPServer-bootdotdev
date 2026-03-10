package main

import (
	"net/http"
)

//This function is a middleware wrapper for our main /app/ endpoint to set up for
//counting the nubmer of times that endpoint has been served
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}