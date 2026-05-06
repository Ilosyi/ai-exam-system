package handlers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Login(t *testing.T) {
	tc := SetupTestContext(t)

	payload := map[string]string{
		"username": "student",
		"password": "student123",
	}

	w := MakeRequest(t, "POST", "/api/auth/login", payload, tc.AuthHandler.Login, nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Register(t *testing.T) {
	tc := SetupTestContext(t)

	payload := map[string]string{
		"username": "newuser",
		"password": "newpass123",
	}

	w := MakeRequest(t, "POST", "/api/auth/register", payload, tc.AuthHandler.Register, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
}
