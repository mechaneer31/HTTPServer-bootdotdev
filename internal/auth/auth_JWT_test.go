package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
    secret := "test-secret"
    userID := "user-123"
	userUUID, _ := uuid.Parse(userID)

    tests := []struct {
        name        string
        tokenFunc   func() string
        secret      string
        expectError bool
        expectID    uuid.UUID
    }{
        {
            name: "valid token",
            tokenFunc: func() string {
                token, _ := MakeJWT(userUUID, secret, time.Hour)
                return token
            },
            secret:      secret,
            expectError: false,
            expectID:    userUUID,
        },
        {
            name: "wrong secret",
            tokenFunc: func() string {
                token, _ := MakeJWT(userUUID, secret, time.Hour)
                return token
            },
            secret:      "wrong-secret",
            expectError: true,
        },
        {
            name: "expired token",
            tokenFunc: func() string {
                token, _ := MakeJWT(userUUID, secret, -time.Second)
                return token
            },
            secret:      secret,
            expectError: true,
        },
        {
            name:        "malformed token",
            tokenFunc:   func() string { return "not.a.valid.token" },
            secret:      secret,
            expectError: true,
        },
    }


	passCount := 0
	failCount := 0

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            token := tt.tokenFunc()
            gotID, err := ValidateJWT(token, tt.secret)
		
			

            if tt.expectError && err == nil {
				failCount++
                t.Errorf("expected error but got none")
                return
            }else 
            if !tt.expectError && err != nil {
				failCount++
                t.Errorf("unexpected error: %v", err)
                return
            } else
            if !tt.expectError && gotID != tt.expectID {
				failCount++
                t.Errorf("expected ID %q, got %q", tt.expectID, gotID)
            } else 
			{
				passCount++
				fmt.Printf("test: %v passed\n", tt.name)
			}

			
        })
    }

	fmt.Printf("%d passed, %d failed\n", passCount, failCount)
}