package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Drag0neUsz/Chirpy/internal/auth"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	t.Log("hashing password:", password, "->", hashed)
	if hashed == "" || hashed == password {
		t.Fatalf("Expected non-empty and different from password, got %s", hashed)
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	password := ""
	_, err := auth.HashPassword(password)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	t.Log("hashing password:", "", "->", err)
}

func TestCheckPasswordHashSuccess(t *testing.T) {
	password := "password"
	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	valid, err := auth.CheckPasswordHash(password, hashed)
	if err != nil {
		t.Fatalf("Error checking password hash: %v", err)
	}
	if !valid {
		t.Fatalf("Expected true, got false")
	}
	t.Log("CheckPasswordHash(password, HashPassword(password)):", password, "->", hashed, "->", valid)
}

func TestCheckPasswordHashFailure(t *testing.T) {
	password := "password"
	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	valid, err := auth.CheckPasswordHash("wrongpassword", hashed)
	if err != nil {
		t.Fatalf("Error checking password hash: %v", err)
	}
	if valid {
		t.Fatalf("Expected false, got true")
	}
	t.Log("CheckPasswordHash(wrongpassword, HashPassword(password)):", "wrongpassword", "->", hashed, "->", valid)
}

func TestCheckPasswordHashInvalidHash(t *testing.T) {
	password := "password"
	_, err := auth.CheckPasswordHash(password, "invalid-hash")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	t.Log("CheckPasswordHash(password, invalid-hash):", password, "->", "invalid-hash", "->", err)
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, "test-secret", 1*time.Hour)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}
	t.Log("MakeJWT(userID, test-secret, 1*time.Hour):", userID, "->", token)
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, "test-secret", 1*time.Hour)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}
	id, err := auth.ValidateJWT(token, "test-secret")
	if err != nil {
		t.Fatalf("Error validating JWT: %v", err)
	}
	if id != userID {
		t.Fatalf("Expected userID, got %s", id)
	}
	t.Log("ValidateJWT(token, test-secret):", id, "->", err)
}

func TestValidateJWTInvalidSecret(t *testing.T) {
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, "test-secret", 1*time.Hour)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}
	id, err := auth.ValidateJWT(token, "invalid-secret")
	if err == nil {
		t.Fatalf("Expecting error, got nil")
	}
	t.Log("ValidateJWT(token, invalid-secret):", id, "->", err)
}

func TestValidateJWTExpired(t *testing.T) {
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, "test-secret", -1*time.Millisecond)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}
	id, err := auth.ValidateJWT(token, "test-secret")
	if err == nil {
		t.Fatalf("Expecting error, got nil")
	}
	t.Log("ValidateJWT(expired_token, test-secret):", id, "->", err)
}

func TestValidateJWTInvalidToken(t *testing.T) {
	id, err := auth.ValidateJWT("invalid-token", "test-secret")
	if err == nil {
		t.Fatalf("Expecting error, got nil")
	}
	t.Log("ValidateJWT(invalid-token, test-secret):", id, "->", err)
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer test-token")
	token, err := auth.GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Error getting bearer token: %v", err)
	}
	if token != "test-token" {
		t.Fatalf("Expected test-token, got %s", token)
	}
	t.Log("GetBearerToken(headers):", token, "->", err)
}

func TestGetBearerTokenNoToken(t *testing.T) {
	headers := http.Header{}
	token, err := auth.GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	t.Log("GetBearerToken(headers):", token, "->", err)
}
