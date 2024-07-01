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
)

var amqpChannel *amqp.Channel
var queueName = "payment_status_queue"
var retryDelay = 30 * time.Second // Retry delay duration

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

	log.Println("Gateway service started on :8080")
	go PollPayments()
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

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

	// Check if payment failed and add to retry queue if necessary
	if resp.StatusCode != http.StatusOK {
		log.Println("Card payment failed, adding to retry queue")
		AddToRetryQueue("card-payment", body)
	}
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

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

func getPaymentType(body []byte) string {
	//payment type is included in the message body as a JSON field
	var message map[string]interface{}
	if err := json.Unmarshal(body, &message); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return ""
	}
	return message["paymentType"].(string)
}

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
