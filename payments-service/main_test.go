package main

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"
	"log"

    _ "github.com/lib/pq"
    "github.com/gorilla/mux"
)


func TestMain(m *testing.M) {
    var err error
    postgresURI := os.Getenv("DATABASE_URI")

    db, err = sql.Open("postgres", postgresURI)
    if err != nil {
        log.Fatal(err)
    }

    exitCode := m.Run()

    os.Exit(exitCode)
}

func setupRouter() *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/payments/initiate", InitiatePayment).Methods("POST")
    r.HandleFunc("/payments/status/{id}", GetPaymentStatus).Methods("GET")
    r.HandleFunc("/payments/send-to-mobile", SendToMobile).Methods("POST")
    return r
}

func TestInitiatePayment(t *testing.T) {
    r := setupRouter()

    paymentRequest := PaymentRequest{
        Amount:        100.0,
        Email:         "test@example.com",
        Location:      "Nairobi",
        Username:      "testuser",
        PaymentMethod: "MPESA",
        Phone:         "0700000000",
        FirstName:     "Test",
        LastName:      "User",
        Reason:        "Test payment",
    }

    jsonValue, _ := json.Marshal(paymentRequest)
    req, _ := http.NewRequest("POST", "/payments/initiate", strings.NewReader(string(jsonValue)))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    if rr.Code != http.StatusAccepted {
        t.Errorf("handler returned wrong status code: got %v, expected %v",
            rr.Code, http.StatusAccepted)
    } else {
		log.Println("TestInitiatePayment: passed")
	}
}

func TestSendToMobile(t *testing.T) {
    r := setupRouter()

    mobilePaymentRequest := MobilePaymentRequest{
        AccountID:    "12345",
        PhoneNumber:  "0700000000",
        Amount:       100.0,
        Narration:    "Test payment",
        CallbackURL:  "https://example.com/callback",
        Channel:      "MPESA",
    }

    jsonValue, _ := json.Marshal(mobilePaymentRequest)
    req, _ := http.NewRequest("POST", "/payments/send-to-mobile", strings.NewReader(string(jsonValue)))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    if rr.Code != http.StatusAccepted {
        t.Errorf("handler returned wrong status code: got %v, expected %v",
            rr.Code, http.StatusAccepted)
    } else {
		log.Println("TestSendToMobile: passed")
	}
}

func TestGetPaymentStatus(t *testing.T) {
    r := setupRouter()

    // Assume a payment with ID 1 exists
    req, _ := http.NewRequest("GET", "/payments/status/1", nil)

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v, expected %v",
            rr.Code, http.StatusOK)
    } else {
		log.Println("TestGetPaymentStatus: passed")
	}
}
