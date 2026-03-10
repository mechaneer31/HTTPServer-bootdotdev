package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Fatal(err)
	}
	
	return hashPassword, nil
}


func CheckPasswordHash(password, hashed_password string) (bool, error) {
	var err error
	var result bool
	
	matchPassword, err := argon2id.ComparePasswordAndHash(password, hashed_password)
	if err != nil {
		log.Println(err)
	}

	if matchPassword == true {
		result = true
		err = nil
		return result, err

	} else {
		result = false
		err = fmt.Errorf("Incorrect password")
		return result, err
	}
	
}

func MakeJWT(userID uuid.UUID,tokenSecret string, expiresIn time.Duration) (string, error) {

	

	key := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims {
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),	
	})

	signature, err := token.SignedString(key)
	if err != nil {
		log.Fatal(err)
	}

	return signature, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	claims := &jwt.RegisteredClaims{}
	
	
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func (token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != "HS256" {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(tokenSecret), nil
		},)
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("Invalid token")
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		log.Fatal(err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Fatal(err)
	}

	return userUUID, nil

}

func GetBearerToken (headers http.Header) (string, error) {

	authHeader := headers.Get("Authorization")
	
	if authHeader == "" {
		err := errors.New("error: authorization not in headers")
		return "", err
	}

	if !strings.HasPrefix(authHeader, "Bearer") {
		err := errors.New("error; authorization header present, no bearer token")
		return "", err
	}

	authToken := strings.TrimPrefix(authHeader, "Bearer ")

	
	return authToken, nil

}


func MakeRefresherToken() string {
	
	numBytes := make([]byte, 32)
	rand.Read(numBytes)
	refreshToken := hex.EncodeToString(numBytes)

	return refreshToken


}

func GetAPIKey(headers http.Header) (string, error){

	authHeader := headers.Get("Authorization")
	
	if authHeader == "" {
		err := errors.New("error: authorization not in headers")
		return "", err
	}

	if !strings.HasPrefix(authHeader, "ApiKey") {
		err := errors.New("error; authorization header present, no poka APIKey")
		return "", err
	}

	authToken := strings.TrimPrefix(authHeader, "ApiKey ")

	
	return authToken, nil


}