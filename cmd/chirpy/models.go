package main

import (
	"net/http"
	"sync/atomic"

	"github.com/mechaneer31/HTTPServer/internal/database"
)

//This struct defines the attributes of a Server instance
type Server struct{
	Handler http.Handler
	Addr 	string
}


//this struct contains the atomic item which is a struct itself
//accessing this location requires pointers, this creates the safety
//of not accidentally overwriting the data
type apiConfig struct{
	fileserverHits 	atomic.Int32
	db				*database.Queries
	pEnv 			string
	jsk				string
	pk				string
}

type errorResponse struct{
	Code		int			`json:"code"`
	ErrMsg		string		`json:"error"`
}



