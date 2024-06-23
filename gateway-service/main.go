package main

import (
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

    log.Println("Gateway service started on :8080")
    http.ListenAndServe(":8080", r)
}

func Register(w http.ResponseWriter, r *http.Request) {
    http.Post("http://localhost:8081/auth/register", "application/json", r.Body)
}

func Login(w http.ResponseWriter, r *http.Request) {
    http.Post("http://localhost:8081/auth/login", "application/json", r.Body)
}

func InitiatePayment(w http.ResponseWriter, r *http.Request) {
    http.Post("http://localhost:8082/payments/initiate", "application/json", r.Body)
}

func GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    http.Get("http://localhost:8082/payments/status/" + id)
}
