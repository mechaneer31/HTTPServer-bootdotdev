package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mechaneer31/HTTPServer/internal/auth"
	"github.com/mechaneer31/HTTPServer/internal/database"
)




func  (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	
	type parameters struct{
		Email 		string 		`json:"email"`
		Password 	string		`json:"password"`
	}	

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Fprintf(w, "Error decoding parameters: %s", err)
	}
	defer r.Body.Close()

	ctx := r.Context()

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Fatal(err)
	}

	createParams := database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(ctx, createParams)
	if err != nil {
		log.Fatal(err)
	}

	jsonSend := databaseInternalUserToUser(user)


	respondWithJSON(w, http.StatusOK, jsonSend)

}

func (cfg *apiConfig) userLoginHandler(w http.ResponseWriter, r *http.Request) {

	
	//Getting the username and password for the user in the request
	type parameters struct{
		Email 		string 				`json:"email"`
		Password 	string 				`json:"password"`
	}

	//Decode the request JSON that has username and password attached
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "Error logging user in, issue with decoding json"
		respondWithError(w, errCode, errMsg)
		return
	}
	defer r.Body.Close()

	//fmt.Printf("params: %v\n", params) //**debug
	//fmt.Printf("params email: %v\n", params.Email) //**debug

		
	//Retrieve user data from database
	ctx := r.Context()
	//Retrieve data based on email passed in from request:
	user, err := cfg.db.GetUserFromEmail(ctx, params.Email)
	//fmt.Printf("user info from database: %v\n", user) //**debug
	//fmt.Println(err) //**debug
	if err != nil {
		//error if user doesn't exist in database (no rows returned by SQL query)
		if err == sql.ErrNoRows {			
			errMsg := "Incorrect email or no user email exists"
			errCode := http.StatusUnauthorized						
			respondWithError(w, errCode, errMsg)
			return
		}

		//Internal Server error when trying to retrieve user data
		httpCode := http.StatusInternalServerError
		err = errors.New("server acting up/get user... ")
		log.Println(httpCode, err)
		return
	}

	//Start user authorization check
	authCheck, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		httpCode := http.StatusInternalServerError
		err = errors.New("server acting up/password check... ")
		log.Println(httpCode, err)
		return
	}

	//Password authorization fails:
	//Password does not match HashedPassword in users Table
	if authCheck == false {
		errMsg := "Incorrect password"
		errCode := http.StatusUnauthorized						
		respondWithError(w, errCode, errMsg)
		return

	} else {

		//Password authentication passed, continue login
		
		//****** JWT ACCESS TOKEN CREATE *******
		//set Access Token (JWT) expiration duration to 1 hour
		accessTokenExp := time.Duration(1) * time.Hour
		//Make JWT access Token from /internal/auth
		jwt, err := auth.MakeJWT(user.ID, cfg.jsk, accessTokenExp)
		if err != nil {
			err = errors.New("failed to make JSON Web Token")
		}


		//********* REFRESH TOKEN CREATE **********
		//Get a refresh made from /internal/auth
		refreshToken := auth.MakeRefresherToken()
		//Set refresh token expiration to 60 days
		refreshTokenExpDuration := time.Duration(1440) * time.Hour
		refreshTokenExp := time.Now().UTC().Add(refreshTokenExpDuration)
		//Sbumit refresh token information to refresh_tokens table:

		refreshTokenDataToCreate := database.CreateRefreshTokenParams{
			Token: refreshToken,
			UserID: user.ID,
			ExpiresAt: refreshTokenExp,
		}
		cfg.db.CreateRefreshToken(ctx, database.CreateRefreshTokenParams(refreshTokenDataToCreate))


		
		//Struct that identifies which information will be returned in the response:
		type UserAuth struct {
			UserID			uuid.UUID	`json:"id"`
			CreatedAt		time.Time	`json:"created_at"`
			UpdatedAt		time.Time	`json:"updated_at"`
			Email			string		`json:"email"`
			IsChirpyRed		bool		`json:"is_chirpy_red"`
			AccessToken		string		`json:"access_token"`
			RefreshToken	string		`json:"refresh_token"`
		}

		respBody:= UserAuth{
			UserID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
			IsChirpyRed: user.IsChirpyRed.Bool,
			AccessToken: jwt,
			RefreshToken: refreshToken,
		}

		respondWithJSON(w, http.StatusOK, respBody)
	}


}

func (cfg *apiConfig) userUpdateInfo(w http.ResponseWriter, r *http.Request){

	
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

	
	type updatedInfoSubmitParameters struct {
		Email		*string		`json:"email,omitempty"`
		Password	*string		`json:"password,omitempty"`
	}


	decoder := json.NewDecoder(r.Body)
	params := (&updatedInfoSubmitParameters{})
	err = decoder.Decode(&params)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "userUpdateInfo Func: error decoding json"
		respondWithError(w, errCode, errMsg)
	}
	defer r.Body.Close()
	

	ctx := r.Context()

	dbUser, _ := cfg.db.GetUserFromID(ctx, validJWT)

	if params.Email != nil && params.Password != nil {
		newHashedPassword, err := auth.HashPassword(*params.Password)
		if err != nil {
			errcode := http.StatusInternalServerError
			errMsg := "Error hashing password in user update"
			respondWithError(w, errcode, errMsg)
		}
		
		updateParams := database.UpdateUserEmailPasswordParams{
			Email: *params.Email,
			HashedPassword: newHashedPassword,
			ID: dbUser.ID,
		}

		cfg.db.UpdateUserEmailPassword(ctx, updateParams)

	} else if params.Email != nil {
		updateParams := database.UpdateUserEmailParams{
			Email: *params.Email,
			ID: dbUser.ID,
		}

		cfg.db.UpdateUserEmail(ctx, updateParams)

	} else if params.Password != nil {
		newHashedPassword, err := auth.HashPassword(*params.Password)
		if err != nil {
			errcode := http.StatusInternalServerError
			errMsg := "Error hashing password in user update"
			respondWithError(w, errcode, errMsg)
		}
		
		updateParams := database.UpdateUserPasswordParams{
			HashedPassword: newHashedPassword,
			ID: dbUser.ID,
		}

		cfg.db.UpdateUserPassword(ctx, updateParams)
	}

	dbUpdatedUser, err := cfg.db.GetUserFromID(ctx, validJWT)
	if err != nil {
		errCode := http.StatusInternalServerError
		errMsg := "Error getting updated user"
		respondWithError(w, errCode, errMsg)
		return
	}

	type userUpdateResponse struct {
		UserID			uuid.UUID	`json:"id"`
		CreatedAt		time.Time	`json:"created_at"`
		UpdatedAt		time.Time	`json:"updated_at"`
		Email			string		`json:"email"`
		IsChirpyRed		bool		`json:"is_chirpy_red"`
		}

	respBody := userUpdateResponse{
		UserID: dbUpdatedUser.ID,
		CreatedAt: dbUpdatedUser.CreatedAt,
		UpdatedAt: dbUpdatedUser.UpdatedAt,
		Email: dbUpdatedUser.Email,
		IsChirpyRed: dbUpdatedUser.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusOK, respBody)


}


func (cfg *apiConfig) addUserSubscription(w http.ResponseWriter, r *http.Request){

	polkaKey, err := auth.GetAPIKey(r.Header)
	if err != nil{
		errCode := http.StatusInternalServerError
		errMsg := "Error extracting polka key from header"
		respondWithError(w, errCode, errMsg)
		return
	}

	if polkaKey != cfg.pk {
		errCode := http.StatusUnauthorized
		errMsg := "Unauthorized to change chirpy user upgrade status"
		respondWithError(w, errCode, errMsg)
		return
	}
	
	type subscriptionRequestParameters struct{
		Event 	string 						`json:"user.upgraded"`
		Data	map[string]uuid.UUID 		`json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := &subscriptionRequestParameters{}
	err = decoder.Decode(&params)
	if err != nil{
		errCode := http.StatusInternalServerError
		errMsg := "error decoding json in updateAddUserSubscription func"
		respondWithError(w, errCode, errMsg)
		return
	}

	defer r.Body.Close()

	ctx := r.Context()

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
	}

	upgradeUserErr := cfg.db.UpdateAddUserSubscription(ctx, params.Data["user_id"])
	if upgradeUserErr != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}