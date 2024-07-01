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
	"github.com/swaggo/http-swagger" 
	_ "github.com/tufstraka/pps/gateway-service/docs" 
)

var amqpChannel *amqp.Channel
var queueName = "payment_status_queue"
var retryDelay = 30 * time.Second 

// @title Payment Gateway API
// @version 1.0
// @description This is a a payment gateway.

// @contact.name API Support
// @contact.email keithkadima@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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
	r.HandleFunc("/register", Register).Methods("POST")
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/payments/initiate", InitiatePayment).Methods("POST")
	r.HandleFunc("/payments/status/{id}", GetPaymentStatus).Methods("GET")
	r.HandleFunc("/payments/send-to-mobile", SendToMobile).Methods("POST")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Gateway service started on :8080")
	go PollPayments()
	http.ListenAndServe(":8080", r)
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user in the system
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Register request"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8081/auth/register", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
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

// Login godoc
// @Summary Log in a user
// @Description Log in a user and return a token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Login request"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8081/auth/login", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to login user: %v", err)
		http.Error(w, "Failed to login user", http.StatusInternalServerError)
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

// InitiatePayment godoc
// @Summary Initiate a new payment
// @Description Initiate a new payment transaction
// @Tags payments
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Payment request"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/initiate [post]
func InitiatePayment(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8082/payments/initiate", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to initiate payment: %v", err)
		http.Error(w, "Failed to initiate payment", http.StatusInternalServerError)
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

	if resp.StatusCode != http.StatusOK {
		log.Println("Card payment failed, adding to retry queue")
		AddToRetryQueue("card-payment", body)
	}
}

// GetPaymentStatus godoc
// @Summary Get payment status
// @Description Get the status of a payment by ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/status/{id} [get]
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
// @Summary Send money to mobile
// @Description Send money to a mobile number
// @Tags payments
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Send to mobile request"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/send-to-mobile [post]
func SendToMobile(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8082/payments/send-to-mobile", "application/json", r.Body)
	if err != nil {
		log.Printf("Failed to send money to mobile: %v", err)
		http.Error(w, "Failed to send money to mobile", http.StatusInternalServerError)
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

	if resp.StatusCode != http.StatusOK {
		log.Println("Mobile payment failed, adding to retry queue")
		AddToRetryQueue("send-to-mobile", body)
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

// PollPayments godoc
// @Summary Poll payments from RabbitMQ queue
// @Description Polls the RabbitMQ queue for payment messages and processes them
// @Tags payments
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal Server Error"
// @Router /poll-payments [get]
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

			if paymentType := getPaymentType(d.Body); paymentType == "send-to-mobile" {
				go RetrySendToMobile(d.Body)
			} else if paymentType == "card-payment" {
				go RetryCardPayment(d.Body)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// getPaymentType determines the payment type from message body
func getPaymentType(body []byte) string {
	//payment type is included in the message body as a JSON field
	var message map[string]interface{}
	if err := json.Unmarshal(body, &message); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return ""
	}
	return message["paymentType"].(string)
}

// RetrySendToMobile retries sending money to mobile
func RetrySendToMobile(body []byte) {
	time.Sleep(retryDelay)
	resp, err := http.Post("http://localhost:8082/payments/send-to-mobile", "application/json", io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		log.Printf("Failed to retry send money to mobile: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Retry failed, will retry again later")
		AddToRetryQueue("send-to-mobile", body)
	} else {
		log.Printf("Retry succeeded")
	}
}

// RetryCardPayment retries initiating card payment
func RetryCardPayment(body []byte) {
	time.Sleep(retryDelay)
	resp, err := http.Post("http://localhost:8082/payments/initiate", "application/json", io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		log.Printf("Failed to retry card payment: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Retry failed, will retry again later")
		AddToRetryQueue("card-payment", body)
	} else {
		log.Printf("Retry succeeded")
	}
}

