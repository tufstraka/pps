package main

import (
    "os"
    "log"
    "net/http"
    "encoding/base64"
    "bytes"
    "database/sql"
    "encoding/json"
    "github.com/joho/godotenv"

    "github.com/gorilla/mux"
    //"github.com/streadway/amqp"
    _ "github.com/lib/pq"
)

var db *sql.DB

func main() {
    godotenv.Load()

    var err error
    postgresURI := os.Getenv("DATABASE_URL")

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
}

func InitiatePayment(w http.ResponseWriter, r *http.Request) {
    var payment PaymentRequest
    err := json.NewDecoder(r.Body).Decode(&payment)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    username := os.Getenv("PAYD_USERNAME")
    password := os.Getenv("PAYD_PASSWORD")

    auth := username + ":" + password
    authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))

    jsonBody, err := json.Marshal(payment)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    paydAPIURL := "https://api.mypayd.app/api/v1/payments"
    req, err := http.NewRequest("POST", paydAPIURL, bytes.NewBuffer(jsonBody))
    if err != nil {
        http.Error(w, "Failed to create request", http.StatusInternalServerError)
        return
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Basic "+authEncoded)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil || resp.StatusCode != http.StatusOK {
        http.Error(w, "Failed to initiate payment", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    _, err = db.Exec("INSERT INTO payments (amount, currency, method, status) VALUES ($1, $2, $3, $4)", payment.Amount, "KES", payment.PaymentMethod, "PENDING")
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    //TO DO: Publish a message to the message queue
    /*conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

    err = ch.Publish(
        "",
        q.Name,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        jsonBody,
        })
    if err != nil {
        log.Fatal(err)
    }*/

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