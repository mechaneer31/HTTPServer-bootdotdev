package main

import (
	"fmt"
	"net/http"
)

//This func identifies the health of the server indicating if the server
//is ready to start receiving requests
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}



//This function is used to show the number of serves for the wrapper function above
func (cfg *apiConfig) metricsHandler (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html><html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!
	</p></body></html>`, cfg.fileserverHits.Load())))
}