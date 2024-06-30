package main

import (
	//"bytes"
	//"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/register", Register).Methods("POST")
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/payments/initiate", InitiatePayment).Methods("POST")
	r.HandleFunc("/payments/status/{id}", GetPaymentStatus).Methods("GET")
	r.HandleFunc("/payments/send-to-mobile", SendToMobile).Methods("POST") 

	log.Println("Gateway service started on :8080")
	http.ListenAndServe(":8080", r)
}

func Register(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8081/auth/register", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

}

func Login(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8081/auth/login", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to login user: %v", err)
		http.Error(w, "Failed to login user", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

}

func InitiatePayment(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8082/payments/initiate", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to initiate payment: %v", err)
		http.Error(w, "Failed to initiate payment", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

}

func GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	resp, err := http.Get("http://localhost:8082/payments/status/" + id)
	if err != nil {
		log.Printf("Failed to get payment status: %v", err)
		http.Error(w, "Failed to get payment status", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

}

func SendToMobile(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8082/payments/send-to-mobile", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to send money to mobile: %v", err)
		http.Error(w, "Failed to send money to mobile", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

}


