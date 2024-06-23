package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "github.com/streadway/amqp"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("postgres", "user=postgres dbname=payment sslmode=disable")
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
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
    Method   string  `json:"method"`
}

func InitiatePayment(w http.ResponseWriter, r *http.Request) {
    var payment PaymentRequest
    err := json.NewDecoder(r.Body).Decode(&payment)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // Store payment record in the database
    _, err = db.Exec("INSERT INTO payments (amount, currency, method, status) VALUES ($1, $2, $3, $4)", payment.Amount, payment.Currency, payment.Method, "PENDING")
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Publish a message to the message queue
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatal(err)
    }
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "payment_queue",
        false,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Serialize the payment struct to JSON
    body, err := json.Marshal(payment)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    err = ch.Publish(
        "",
        q.Name,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
    if err != nil {
        log.Fatal(err)
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
