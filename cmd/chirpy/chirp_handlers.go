package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mechaneer31/HTTPServer/internal/auth"
	"github.com/mechaneer31/HTTPServer/internal/database"
)

//This function validates the size of a chirp
func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	
	
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("tokenString: %v", tokenString)

	jwtUser, err := auth.ValidateJWT(tokenString, cfg.jsk)
	if err != nil {
		fmt.Println(err)
		errMsg := fmt.Errorf("token not validated")
		respondWithError(w, http.StatusUnauthorized, errMsg.Error())
		return
	}
	
	
	fmt.Printf("in chirpHandler func...\n")
	
	type parameters struct{
		ChirpMessage	string		`json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		fmt.Fprintf(w, "Error decoding parameters: %s", err)
	}
	defer r.Body.Close()


	fmt.Printf("chirp decoded...\n")

	charCount := utf8.RuneCountInString(params.ChirpMessage)	
	if charCount > 140 {		

		err = fmt.Errorf("Chirp is too long")

		respBody := errorResponse{
			Code: http.StatusBadRequest,
			ErrMsg: err.Error(),
		}
		fmt.Printf("chirp failed because too long...\n")
		respondWithError(w, respBody.Code, respBody.ErrMsg)
		

	} else {
		fmt.Printf("building chirp json...\n")
		removeProfanity(&params.ChirpMessage)		

		fmt.Printf("profanity removed...\n")

		chirpParameters := database.CreateChirpParams{
			Body: params.ChirpMessage,
			UserID: jwtUser,
		}
		
		ctx := r.Context()
		dbChirp, err := cfg.db.CreateChirp(ctx, chirpParameters)
		if err != nil{
			fmt.Printf("error at creating chirp to database...\n")
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		newChirp := databaseExternalChirpToChirp(dbChirp)

		
		fmt.Printf("new chirp created and saved to database, sending json...\n")

		respondWithJSON(w, http.StatusCreated, newChirp)

	}
}

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request){

	ctx := r.Context()
	
	optSingleUserChirpsQuery := r.URL.Query().Get("author_id")
	fmt.Printf("optQuery: %v\n", optSingleUserChirpsQuery)

	optSortingQuery := r.URL.Query().Get("sort")

	if optSingleUserChirpsQuery != "" {
		parseUUID, err := uuid.Parse(optSingleUserChirpsQuery)
		if err != nil{
			errCode := http.StatusInternalServerError
			errMsg := "error parsing UUID from optQuery in getAllChirpsHandler func"
			respondWithError(w, errCode, errMsg)
			return
		}

		dbChirps, err := cfg.db.GetAllChirpsByID(ctx, parseUUID)
		if err != nil {
			errCode := http.StatusInternalServerError
			errMsg := "error at GetAllChirpsByID func in getAllChirpsHandler func"
			respondWithError(w, errCode, errMsg)
			return
		}

		respChirps := databseExternalChirpstoChirps(dbChirps)

		if optSortingQuery == "desc" {
			sort.Slice(respChirps, func(i, j int) bool {
				return respChirps[i].CreatedAt.After(respChirps[j].CreatedAt) 
			})

			respondWithJSON(w, http.StatusOK, respChirps)

		} else {

			respondWithJSON(w, http.StatusOK, respChirps)
		}

		



	} else {

		dbChirps, err := cfg.db.GetAllChirps(ctx)
		if err != nil {
			errCode := http.StatusInternalServerError
			errMsg := "error at GetAllChirpsByID func in getAllChirpsHandler func"
			respondWithError(w, errCode, errMsg)
			return
		}



		respChirps := databseExternalChirpstoChirps(dbChirps)		

		if optSortingQuery == "desc" {
			sort.Slice(respChirps, func(i, j int) bool {
				return respChirps[i].CreatedAt.After(respChirps[j].CreatedAt) 
			})

			respondWithJSON(w, http.StatusOK, respChirps)

		} else {

			respondWithJSON(w, http.StatusOK, respChirps)
		}
		
		
	}
	
	

}

func (cfg *apiConfig) getSingleChirpHandler(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Path	uuid.UUID  `json:"chirpID"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Fprintf(w, "Error decoding parameters: %s", err)
	}
	defer r.Body.Close()

	path := r.PathValue("chirpID")
	fmt.Printf("path: %v\n", path)

	if path == "" {
		errMsg := fmt.Errorf("Chirp does not exist")
		respondWithError(w, http.StatusNotFound, errMsg.Error())
	} else {

		ctx := r.Context()
		dbChirp, err := cfg.db.GetSingleChirp(ctx, params.Path)
		if err != nil {
			fmt.Printf("error at getting single chirp from db...\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respChirps := databaseExternalChirpToChirp(dbChirp)

		respondWithJSON(w, http.StatusOK, respChirps)

	}

}



func (cfg *apiConfig) deleteChirpByID(w http.ResponseWriter, r *http.Request){

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errCode := http.StatusUnauthorized
		errMsg := "Issue with getting Bearer token"
		respondWithError(w, errCode, errMsg)
		return
	}

	validJWT, err := auth.ValidateJWT(tokenString, cfg.jsk)
	if err != nil {
		errCode := http.StatusUnauthorized
		errMsg := "Issue validating JWT"
		respondWithError(w, errCode, errMsg)
	}

	type chirpDeleteParams struct{
		ChirpID 	string		`json:"chirp_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := &chirpDeleteParams{}
	err = decoder.Decode(&params)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "Issue with decoding json to delete chirp"
		respondWithError(w, errCode, errMsg)
		return
	}

	defer r.Body.Close()

	parseChirpID, err := uuid.Parse(params.ChirpID)
	if err != nil{
		errCode := http.StatusInternalServerError
		errMsg := "error parsing uuid from string in deleteChripByID func"
		respondWithError(w, errCode, errMsg)
		return
	}

	ctx := r.Context()
	dbChirp, err := cfg.db.GetSingleChirp(ctx, parseChirpID)
	if err != nil {
		errCode := http.StatusNotFound
		errMsg := "chirp was not found in database"
		respondWithError(w, errCode, errMsg)
		return
	}
	fmt.Printf("dbChirp: %v\n", dbChirp)

	chirpInfo := databaseInternalChirpToChirp(dbChirp)
	fmt.Printf("chirpInfo: %v\n", chirpInfo)
	fmt.Printf("validJWT: %v\n", validJWT)

	if chirpInfo.UserID != validJWT {
		errCode := http.StatusForbidden
		errMsg := "You are not authorized to delete this chirp"
		respondWithError(w, errCode, errMsg)
		return
	}

	fmt.Printf("chirp to delete by id: %v\n", chirpInfo.ID)

	cfg.db.DeleteSingleChirp(ctx, chirpInfo.ID)
	w.WriteHeader(http.StatusNoContent)


}


func removeProfanity(message *string) {
	tempMess := strings.Split(*message, " ")	

	for i, word := range tempMess {
	tempWord := strings.ToLower(word)

		if tempWord == "kerfuffle" {
			tempMess[i] = "****"			
		}
		
		if tempWord == "sharbert" {
			tempMess[i] = "****"			
		}

		if tempWord == "fornax" {
			tempMess[i] = "****"			
		}
	}	
	
	*message = strings.Join(tempMess, " ")
	
}