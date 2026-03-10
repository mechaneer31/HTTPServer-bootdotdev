package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mechaneer31/HTTPServer/internal/auth"
	"github.com/mechaneer31/HTTPServer/internal/database"
)


func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request){
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "refreshToken Func: Error getting Bearer token from request header"
		respondWithError(w, errCode, errMsg)
		return
	}

	ctx := r.Context()
	dbRefreshToken, err := cfg.db.GetUserFromRefreshToken(ctx, refreshToken)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "refreshToken Func: error getting user info from refresh_tokens database"
		respondWithError(w, errCode, errMsg)
		return
	}

	currentTime := time.Now().UTC()

	if dbRefreshToken.Token == "" || 
		dbRefreshToken.ExpiresAt.UTC().Before(currentTime) ||
		dbRefreshToken.RevokedAt.Valid  {
			errCode := http.StatusUnauthorized
			errMsg := "Unauthorized due to expired/revoked refresh token"
			respondWithError(w, errCode, errMsg)
			return
		} else {
			accessTokenExp := time.Duration(1) * time.Hour
			newAccessToken, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.jsk, accessTokenExp)
			if err != nil {
				errCode := http.StatusInternalServerError
				errMsg := "Error creating jwt token in refresh_token function"
				respondWithError(w, errCode, errMsg)
				return
			}
			respondWithJSON(w, http.StatusOK, newAccessToken)
		}

}


func (cfg *apiConfig) revokeRefreshToken(w http.ResponseWriter, r *http.Request){

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "revokeRefreshToken Func: Error getting Bearer token from request header"
		respondWithError(w, errCode, errMsg)
		return
	}

	ctx := r.Context()
	dbRefreshToken, err := cfg.db.GetUserFromRefreshToken(ctx, refreshToken)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "revokeRefreshToken Func: error getting user info from refresh_tokens database"
		respondWithError(w, errCode, errMsg)
		return
	}

	dbParameters := database.RevokeRefreshTokenParams{
		Token: refreshToken,
		UserID: dbRefreshToken.UserID,
	}

	cfg.db.RevokeRefreshToken(ctx, dbParameters)

	w.WriteHeader(http.StatusNoContent)

}


//This function allows us to reset the number of serve hits to 0 stored in the atomic var
func (cfg *apiConfig) resetHandler (w http.ResponseWriter, r *http.Request) {
	if cfg.pEnv != "dev"{
		w.WriteHeader(http.StatusForbidden)
		return
	}
	
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "counter reset\n")

	ctx := r.Context()
	cfg.db.DeleteAllUsers(ctx)

}


func respondWithError(w http.ResponseWriter, code int, msg string) {

	dat, err := json.Marshal(msg)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}


func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){

	dat, err := json.Marshal(payload)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}


func main(){

godotenv.Load(".env")


dbURL := os.Getenv("DB_URL")
platform := os.Getenv("PLATFORM")
jwt_secret_key := os.Getenv("JWT_SECRET_KEY")
polka_key := os.Getenv("POLKA_KEY")

fmt.Println(os.Getenv("DB_URL"))

db, err := sql.Open("postgres", dbURL)
if err != nil {
	log.Fatal(err)
}

defer db.Close()

dbQueries := database.New(db)


//Initialize an instance of apiConfig which initializes the atomic struct
cfg := apiConfig{
	db: dbQueries,
	pEnv: platform,
	jsk : jwt_secret_key,
	pk : polka_key,
}

//Creates a new ServeMux which is a HTTP request multiplexer (fancy words for router).
//Each mux below is a mux pattern and the ServeMux matches URLs to requests (closest matching)
//to determine where to route the response to. 
mux := http.NewServeMux()	


//Initializing a Server.  This server calls the mux handler such that the
//mux calls below execute.
myServer := Server{
	Handler: mux,
	Addr: ":8080",
}

// mux patterns available for the server to try and match requests to
mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

mux.HandleFunc("GET /api/healthz", readinessHandler)
mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
mux.HandleFunc("POST /api/refresh", cfg.refreshToken)
mux.HandleFunc("POST /api/revoke", cfg.revokeRefreshToken)

//       USER FUNCs
mux.HandleFunc("POST /api/users", cfg.createUserHandler)
mux.HandleFunc("POST /api/login", cfg.userLoginHandler)
mux.HandleFunc("PUT /api/users", cfg.userUpdateInfo)
mux.HandleFunc("POST /api/polka/webhooks", cfg.addUserSubscription)

//       CHIRP FUNCs
mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getSingleChirpHandler)
mux.HandleFunc("GET /api/chirps", cfg.getAllChirpsHandler)
mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpByID)




//This opens the server channel to listen for requests and then the mux
//will find a match for sending the response
http.ListenAndServe(myServer.Addr, myServer.Handler)

}  //main function end