package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    "github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
    payload := `{"username":"testuser","password":"password"}`
    req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(payload))
    response := httptest.NewRecorder()
    Register(response, req)

    assert.Equal(t, http.StatusCreated, response.Code)
}
