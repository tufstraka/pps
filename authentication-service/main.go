package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"
    "fmt"

    "github.com/joho/godotenv"

    "github.com/dgrijalva/jwt-go"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "golang.org/x/crypto/bcrypt"

    // Swagger dependencies
    _ "github.com/tufstraka/pps/authentication-service/docs" 
    "github.com/swaggo/http-swagger"
)

// @title Authentication Service API
// @version 0.1
// @description This is an authentication service with JWT.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email keithkadima@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8085
// @BasePath /

var db *sql.DB

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    postgresURI := os.Getenv("DATABASE_URI")
    log.Printf("Connecting to database with URI: %s", postgresURI)

    db, err = sql.Open("postgres", postgresURI)
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatalf("Error connecting to the database: %v", err)
    }

    r := mux.NewRouter()
    r.HandleFunc("/auth/register", Register).Methods("POST")
    r.HandleFunc("/auth/login", Login).Methods("POST")

    r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

    log.Println("Authentication service started on :8085")
    err = http.ListenAndServe(":8085", r)
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
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

type Claims struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Location string `json:"location"`
    Phone    string `json:"phone"`
    jwt.StandardClaims
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with the provided details
// @Tags auth
// @Accept  json
// @Produce  json
// @Param user body User true "User Details"
// @Success 201 {string} string "Created"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/register [post]
func Register(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        log.Printf("Error decoding JSON: %v", err)
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("Error generating password hash: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    _, err = db.Exec("INSERT INTO users (username, password_hash, location, phone, email) VALUES ($1, $2, $3, $4, $5)",
        user.Username, string(hashedPassword), user.Location, user.Phone, user.Email)
    if err != nil {
        log.Printf("Error executing insert: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

// Login godoc
// @Summary Login a user
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body UserLogin true "User Details"
// @Success 200 {string} string "OK"
// @Failure 401 {string} string "Invalid Credentials"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
    var user User
    my_secret_key := os.Getenv("JWT_SECRET_KEY")
    var jwtKey = []byte(my_secret_key)
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        log.Printf("Error decoding JSON: %v", err)
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    var storedHash string
    err = db.QueryRow("SELECT password_hash FROM users WHERE username=$1", user.Username).Scan(&storedHash)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
            return
        }
        log.Printf("Error querying database: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(user.Password))
    if err != nil {
        log.Printf("Error comparing password hash: %v", err)
        http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
        return
    }

    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        Username: user.Username,
        Email: user.Email,
        Location: user.Location,
        Phone: user.Phone,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        log.Printf("Error signing token: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:    "token",
        Value:   tokenString,
        Expires: expirationTime,
    })

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Login successful")


}