package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/tufstraka/pps/payments-service/docs"
)

var db *sql.DB

// @title Payment APIs
// @version 0.1
// @description This is a payment service with Payd API integration.

// @contact.name API Support
// @contact.email keithkadima@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /

func main() {
	godotenv.Load()

	var err error
	postgresURI := os.Getenv("DATABASE_URI")

	db, err = sql.Open("postgres", postgresURI)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/payments/initiate", InitiatePayment).Methods("POST")
	r.HandleFunc("/payments/status/{id}", GetPaymentStatus).Methods("GET")
	r.HandleFunc("/payments/send-to-mobile", SendToMobile).Methods("POST")
	r.HandleFunc("/payments/get-card-details", GetCardDetails).Methods("POST")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Payments service started on :8082")
	http.ListenAndServe(":8082", r)
}

type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	Email         string  `json:"email"`
	Location      string  `json:"location"`
	Username      string  `json:"username"`
	PaymentMethod string  `json:"payment_method"`
	Phone         string  `json:"phone"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Reason        string  `json:"reason"`
}

type MobilePaymentRequest struct {
	AccountID     string  `json:"account_id"`
	PhoneNumber   string  `json:"phone_number"`
	Amount        float64 `json:"amount"`
	Narration     string  `json:"narration"`
	CallbackURL   string  `json:"callback_url"`
	Channel       string  `json:"channel"`
	PaymentMethod string  `json:"payment_method"`
}

// InitiatePayment godoc
// @Summary Initiate a payment
// @Description Initiate a payment to a user
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body PaymentRequest true "Payment Request"
// @Success 202 {string} string "Accepted"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /payments/initiate [post]
func InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var payment PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		http.Error(w, "Bad Request: invalid JSON structure", http.StatusBadRequest)
		return
	}

	var id int
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", payment.Username).Scan(&id)
	if err != nil {
		log.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	username := os.Getenv("PAYD_USERNAME")
	password := os.Getenv("PAYD_PASSWORD")
	auth := username + ":" + password
	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	jsonBody, err := json.Marshal(payment)
	if err != nil {
		http.Error(w, "Internal Server Error: failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	paydAPIURL := "https://api.mypayd.app/api/v1/payments"

	log.Println("Sending payment request to Payd API:")
	log.Printf("Body: %s", jsonBody)
	log.Printf("Authorization: Basic %s", authEncoded)

	req, err := http.NewRequest("POST", paydAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+authEncoded)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		http.Error(w, "Failed to initiate payment", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	log.Printf("Response Status Code: %d", resp.StatusCode)
	log.Printf("Response Body: %s", respBody)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to initiate payment. Status code: %d", resp.StatusCode)
		http.Error(w, string(respBody), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO payments (amount, currency, method, status, user_id) VALUES ($1, $2, $3, $4, $5)",
		payment.Amount, "KES", payment.PaymentMethod, "PENDING", id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// SendToMobile godoc
// @Summary Send money to a mobile number
// @Description Send money to a mobile number via the Payd API
// @Tags payments
// @Accept json
// @Produce json
// @Param mobilePayment body MobilePaymentRequest true "Mobile Payment Request"
// @Success 202 {string} string "Accepted"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /payments/send-to-mobile [post]
func SendToMobile(w http.ResponseWriter, r *http.Request) {
	var mobilePayment MobilePaymentRequest
	err := json.NewDecoder(r.Body).Decode(&mobilePayment)
	if err != nil {
		http.Error(w, "Bad Request: invalid JSON structure", http.StatusBadRequest)
		return
	}

	username := os.Getenv("PAYD_USERNAME")
	password := os.Getenv("PAYD_PASSWORD")
	auth := username + ":" + password
	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	jsonBody, err := json.Marshal(mobilePayment)
	if err != nil {
		http.Error(w, "Internal Server Error: failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	paydAPIURL := "https://api.mypayd.app/api/v2/withdrawal"

	log.Println("Sending mobile payment request to Payd API:")
	log.Printf("Body: %s", jsonBody)
	log.Printf("Authorization: Basic %s", authEncoded)

	req, err := http.NewRequest("POST", paydAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+authEncoded)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		http.Error(w, "Failed to send mobile payment", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	log.Printf("Response Status Code: %d", resp.StatusCode)
	log.Printf("Response Body: %s", respBody)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send mobile payment. Status code: %d, body: %s", resp.StatusCode, respBody)
		http.Error(w, "Failed to send mobile payment: "+string(respBody), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetPaymentStatus godoc
// @Summary Get payment status
// @Description Get the status of a payment by ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} map[string]string
// @Failure 404 {string} string "Payment Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /payments/status/{id} [get]
func GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var status string
	err := db.QueryRow("SELECT status FROM payments WHERE id=$1", id).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Payment Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func GetCardDetails(w http.ResponseWriter, r *http.Request) {
	username := os.Getenv("PAYD_USERNAME")
	password := os.Getenv("PAYD_PASSWORD")
	auth := username + ":" + password
	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	paydAPIURL := "https://api.mypayd.app/api/v2/payments"

	log.Println("Getting card details from Payd API:")
	log.Printf("Authorization: Basic %s", authEncoded)

	req, err := http.NewRequest("POST", paydAPIURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+authEncoded)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	log.Printf("Response Status Code: %d", resp.StatusCode)
	log.Printf("Response Body: %s", respBody)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to initiate payment. Status code: %d", resp.StatusCode)
		http.Error(w, "Failed to initiate payment: "+string(respBody), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
