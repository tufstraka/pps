package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/tufstraka/pps/gateway-service/docs"
)

var amqpChannel *amqp.Channel
var queueName = "payment_status_queue"
var retryDelay = 30 * time.Second

// @title Payment Gateway API
// @version 0.1
// @description This is a payment gateway service that integrates with the authentication and payments services.
// @contact.name API Support
// @contact.email keithkadima@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 54.145.134.156:8083
// @BasePath /

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	amqpChannel, err = conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer amqpChannel.Close()

	_, err = amqpChannel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	r := mux.NewRouter()

	r.Use(loggingMiddleware)

	r.HandleFunc("/register", Register).Methods("POST")
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/payments/initiate", InitiatePayment).Methods("POST")
	r.HandleFunc("/payments/status/{id}", GetPaymentStatus).Methods("GET")
	r.HandleFunc("/payments/send-to-mobile", SendToMobile).Methods("POST")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Gateway service started on :8083")
	go PollPayments()
	http.ListenAndServe(":8083", r)
}

// Middleware function to log request details
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("Started %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)

		if r.Method == "POST" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Failed to read request body: %v", err)
			} else {
				log.Printf("Body: %s", body)
				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset the request body
			}
		}

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("Completed %s %s in %v", r.Method, r.RequestURI, duration)
	})
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Location string `json:"location"`
	Phone    string `json:"phone"`
}

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

type LoginSuccessResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

type SuccessResponse struct {
	User string `json:"user"`
}

type FailResponse struct {
	Status string `json:"status"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with the provided details
// @Tags auth
// @Accept json
// @Produce json
// @Param user body User true "User Details"
// @Success 201 {object} SuccessResponse "Registration successful"
// @Failure 401 {object} FailResponse "Registration failed"
// @Failure 500 {object} FailResponse "Server error"
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://54.145.134.156:8085/auth/register", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		http.Error(w, `{"status": "server error"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		http.Error(w, `{"status": "server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	switch resp.StatusCode {
	case http.StatusOK:
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	case http.StatusUnauthorized:
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status": "invalid credentials"}`))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": "server error"}`))
	}
}


// Login godoc
// @Summary Login a user
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body UserLogin true "User details"
// @Success 200 {object} LoginSuccessResponse "Login successful with user details and token"
// @Failure 401 {object} FailResponse "Invalid credentials"
// @Failure 500 {object} FailResponse "Server error"
// @Router /login [post]
func Login(w http.ResponseWriter, r *http.Request) {
    resp, err := http.Post("http://54.145.134.156:8085/auth/login", "application/json", r.Body)
    if err != nil {
        log.Printf("Failed to login user: %v", err)
        http.Error(w, `{"status": "server error"}`, http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    var response struct {
        Status string `json:"status"`
        Token  string `json:"token,omitempty"`
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Failed to read response: %v", err)
        http.Error(w, `{"status": "server error"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    switch resp.StatusCode {
    case http.StatusOK:
        json.Unmarshal(body, &response)
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    case http.StatusUnauthorized:
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(FailResponse{Status: "invalid credentials"})
    default:
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(FailResponse{Status: "server error"})
    }
}


// InitiatePayment godoc
// @Summary Initiate a payment
// @Description Initiate a payment for a user
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body PaymentRequest true "Payment Request"
// @Success 202 {object} string "Success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /payments/initiate [post]
func InitiatePayment(w http.ResponseWriter, r *http.Request) {
	// Read and store the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://54.145.134.156:8082/payments/initiate", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("Failed to initiate payment: %v", err)
		http.Error(w, "Failed to initiate payment", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)

	if resp.StatusCode != http.StatusCreated {
		log.Println("Card payment failed, adding to retry queue")
		AddToRetryQueue("card-payment", bodyBytes)
	}
}

// GetPaymentStatus godoc
// @Summary Get payment status
// @Description Get the status of a payment by ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {string} string "Accepted"
// @Failure 500 {string} string "Internal Server Error"
// @Router /payments/status/{id} [get]
func GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	resp, err := http.Get("http://54.145.134.156:8082/payments/status/" + id)
	if err != nil {
		log.Printf("Failed to get payment status: %v", err)
		http.Error(w, "Failed to get payment status", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
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
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://54.145.134.156:8082/payments/send-to-mobile", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("Failed to send money to mobile: %v", err)
		http.Error(w, "Failed to send money to mobile", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)

	if resp.StatusCode != http.StatusOK {
		log.Println("Mobile payment failed, adding to retry queue")
		AddToRetryQueue("send-to-mobile", bodyBytes)
	}
}

func AddToRetryQueue(paymentType string, body []byte) {
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	err := amqpChannel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		msg,
	)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
	}
}

func PollPayments() {
	for {
		msgs, err := amqpChannel.Consume(
			queueName, // queue
			"",        // consumer
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		if err != nil {
			log.Fatalf("Failed to register a consumer: %v", err)
		}

		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			if paymentType := getPaymentType(d.Body); paymentType == "mobile" {
				go RetrySendToMobile(d.Body)
			} else if paymentType == "card" {
				go RetryCardPayment(d.Body)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func getPaymentType(body []byte) string {
	var message map[string]interface{}
	if err := json.Unmarshal(body, &message); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return ""
	}

	paymentType, ok := message["payment_method"].(string)
	if !ok {
		log.Println("paymentType not found or is not a string")
		return ""
	}

	return paymentType
}

func RetryCardPayment(body []byte) {
	attempts := 0
	for attempts < 5 {
		time.Sleep(retryDelay)
		resp, err := http.Post("http://54.145.134.156:8082/payments/initiate", "application/json", io.NopCloser(bytes.NewBuffer(body)))
		if err != nil {
			log.Printf("Failed to retry card payment: %v", err)
			attempts++
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Retry failed, will retry again later")
			attempts++
			continue
		} else {
			log.Printf("Retry succeeded")
			return
		}
	}

	log.Printf("Retry attempts exhausted for initiating card payment")
	AddToRetryQueue("card-payment", body)
}

func RetrySendToMobile(body []byte) {
	attempts := 0
	for attempts < 5 {
		time.Sleep(retryDelay)
		resp, err := http.Post("http://54.145.134.156:8082/payments/send-to-mobile", "application/json", io.NopCloser(bytes.NewBuffer(body)))
		if err != nil {
			log.Printf("Failed to retry send money to mobile: %v", err)
			attempts++
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Retry failed, will retry again later")
			attempts++
			continue
		} else {
			log.Printf("Retry succeeded")
			return
		}
	}

	log.Printf("Retry attempts exhausted for sending money to mobile")
	AddToRetryQueue("send-to-mobile", body)
}
