package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/mechaneer31/HTTPServer/internal/database"
)

//********** LOOP TO GET USER DATA FOR INTERNAL OR ADMIN PURPOSES

type InternalUserData struct{
	ID			uuid.UUID		`json:"id"`
	Created_at	time.Time		`json:"created_at"`
	Updated_at 	time.Time		`json:"updated_at"`
	Email		string			`json:"email"`
}

// use this loop to return a single item
func databaseInternalUserToUser(dbUser database.User) InternalUserData {

	return InternalUserData{
		ID: dbUser.ID,
		Created_at: dbUser.CreatedAt,
		Updated_at: dbUser.UpdatedAt,
		Email: dbUser.Email,
	}
}

// use this loop to return multiple items
func databseInternalUsersToUsers(dbUsers []database.User) []InternalUserData{
	internalUsers := make([]InternalUserData, 0, len(dbUsers))

	for _, dbUser := range dbUsers{
		internalUsers = append(internalUsers, databaseInternalUserToUser(dbUser))
	}
	return internalUsers
}//**********   END GET INTERNAL USERS LOOP



//******************   LOOP TO GET USER DATA FOR EXTERNAL OR PUBLIC VIEW PURPOSES
type ExternalUserData struct{
	ID			uuid.UUID		`json:"id"`
	Created_at	time.Time		`json:"created_at"`
	Updated_at 	time.Time		`json:"updated_at"`
	Email		string			`json:"email"`
}

//********** loop to get single item
func databaseExternalUserToUser(dbUser database.User) ExternalUserData {

	return ExternalUserData{
		ID: dbUser.ID,
		Created_at: dbUser.CreatedAt,
		Updated_at: dbUser.UpdatedAt,
		Email: dbUser.Email,
	}
}

//********* loop to get multiple items
func databseExternalUsersToUsers(dbUsers []database.User) []ExternalUserData{
	externalUsers := make([]ExternalUserData, 0, len(dbUsers))

	for _, dbUser := range dbUsers{
		externalUsers = append(externalUsers, databaseExternalUserToUser(dbUser))
	}
	return externalUsers
}//**********   END GET EXTERNAL USERS LOOP