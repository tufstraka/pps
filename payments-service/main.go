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
)

var db *sql.DB

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

func InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var payment PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		http.Error(w, "Bad Request: invalid JSON structure", http.StatusBadRequest)
		return
	}

	// Retrieve user ID based on username
	var id int
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", payment.Username).Scan(&id)
	if err != nil {
		log.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Prepare payment payload for Payd API
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

	// Log request details
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

	// Make the request to external API
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

	// Log response status and body
	log.Printf("Response Status Code: %d", resp.StatusCode)
	log.Printf("Response Body: %s", respBody)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to initiate payment. Status code: %d", resp.StatusCode)
		http.Error(w, "Failed to initiate payment: "+string(respBody), http.StatusBadRequest)
		return
	}

	// Insert payment details into the payments table with user ID
	_, err = db.Exec("INSERT INTO payments (amount, currency, method, status, user_id) VALUES ($1, $2, $3, $4, $5)",
		payment.Amount, "KES", payment.PaymentMethod, "PENDING", id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

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


