package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func TestMain(m *testing.M) {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	postgresURI := os.Getenv("DATABASE_URI")

	db, err = sql.Open("postgres", postgresURI)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v db: %s", err, postgresURI)
	}

	// Clean up the test user before running tests
	_, _ = db.Exec("DELETE FROM users WHERE username=$1", "testuser")

	code := m.Run()

	// Clean up the test user after running tests
	_, _ = db.Exec("DELETE FROM users WHERE username=$1", "testuser")

	db.Close()

	os.Exit(code)
}

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/register", Register).Methods("POST")
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/payments/initiate", InitiatePayment).Methods("POST")
	return r
}

    //==============================================================================================
    // Tests - DO NOT MODIFY
    //==============================================================================================

func TestRegister(t *testing.T) {
	r := setupRouter()

	user := User{
		Username: "testuser",
		Password: "password",
		Email:    "test@example.com",
		Location: "Test City",
		Phone:    "1234567890",
	}

	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusCreated)
	} else {
		log.Println("TestRegister: passed")
	}
}

func TestLogin(t *testing.T) {
	r := setupRouter()

	// Check if the test user already exists in the database
	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username=$1", "testuser").Scan(&storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("Test user not found in the database")
		} else {
			t.Fatalf("Error querying database: %v", err)
		}
	}

	// Compare the stored hash with the expected password
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("password"))
	if err != nil {
		t.Fatalf("Stored password hash does not match the expected password: %v", err)
	}

	user := User{
		Username: "testuser",
		Password: "password",
	}

	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	} else {
		log.Println("TestLogin: passed")
	}
}

func TestInitiatePayment(t *testing.T) {
	// Create a request body for payment initiation
	payment := PaymentRequest{
		Amount:        100.0,
		Email:         "testuser@example.com",
		Location:      "Test City",
		Username:      "testuser",
		PaymentMethod: "card",
		Phone:         "+1234567890",
		FirstName:     "Test",
		LastName:      "User",
		Reason:        "Testing payment",
	}
	body, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Create a request with JSON body
	req, err := http.NewRequest("POST", "/payments/initiate", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the request to the registered handler function
	http.HandlerFunc(InitiatePayment).ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := "Payment initiated successfully"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
