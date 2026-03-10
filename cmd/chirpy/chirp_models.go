package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/mechaneer31/HTTPServer/internal/database"
)

//******* LOOP TO GET CHIRP DATA FOR INTERNAL/ADMIN PURPSOSES

type InternalChirpData struct{
	ID			uuid.UUID		`json:"chirp_id"`
	CreatedAt	time.Time		`json:"created_at"`
	UpdatedAt	time.Time		`json:"updated_at"`
	Body 		string			`json:"body"`
	UserID		uuid.UUID 		`json:"user_id"`
}


func databaseInternalChirpToChirp(dbChirp database.Chirp) InternalChirpData{
	
	return InternalChirpData{
		ID: dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body: dbChirp.Body,
		UserID: dbChirp.UserID,
	}

	
}

func databseInternalChirpstoChirps(dbChirps []database.Chirp) []InternalChirpData{

	internalChirps := make([]InternalChirpData, 0, len(dbChirps))

	for _, dbChirp := range dbChirps{
		internalChirps = append(internalChirps, databaseInternalChirpToChirp(dbChirp))
	}
	return internalChirps
} //************ END GET INTERNAL CHIRPS LOOP



//******* LOOP TO GET CHIRP DATA FOR EXTERNAL/PUBLIC PURPOSES

type ExternalChirpData struct{
	Body 		string			`json:"body"`
	UserID		uuid.UUID 		`json:"user_id"`
	CreatedAt	time.Time		`json:"created_at"`
}


func databaseExternalChirpToChirp(dbChirp database.Chirp) ExternalChirpData{
	

	return ExternalChirpData{
		Body: dbChirp.Body,
		UserID: dbChirp.UserID,
		CreatedAt: dbChirp.CreatedAt,
	}
}

func databseExternalChirpstoChirps(dbChirps []database.Chirp) []ExternalChirpData{

	externalChirps := make([]ExternalChirpData, 0, len(dbChirps))

	for _, dbChirp := range dbChirps{
		externalChirps = append(externalChirps, databaseExternalChirpToChirp(dbChirp))
	}
	return externalChirps
} //************ END GET INTERNAL CHIRPS LOOP