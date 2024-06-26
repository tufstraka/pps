package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    //"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
    // Prepare the request payload
    payload := `{"username":"testuser","password":"password","email":"test@example.com","location":"Test City","phone":"1234567890"}`
    req, err := http.NewRequest("POST", "/auth/register", strings.NewReader(payload))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")

    // Create a ResponseRecorder to record the response
    rr := httptest.NewRecorder()

    // Call the Register handler function directly
    Register(rr, req)

    // Check the status code of the response
    if rr.Code != http.StatusCreated {
        t.Errorf("handler returned wrong status code: got %v want %v",
            rr.Code, http.StatusCreated)
    }

    // You can optionally check the response body for further validation
    // Example:
    // expectedBody := `User successfully registered`
    // assert.Equal(t, expectedBody, rr.Body.String())
}

func TestLogin(t *testing.T) {
    // Prepare the request payload
    payload := `{"username":"testuser","password":"password"}`
    req, err := http.NewRequest("POST", "/auth/login", strings.NewReader(payload))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")

    // Create a ResponseRecorder to record the response
    rr := httptest.NewRecorder()

    // Call the Login handler function directly
    Login(rr, req)

    // Check the status code of the response
    if rr.Code != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            rr.Code, http.StatusOK)
    }

    // You can optionally check the response body for further validation
    // Example:
    // expectedBody := `{"token":"your_generated_jwt_token"}`
    // assert.Equal(t, expectedBody, rr.Body.String())
}
