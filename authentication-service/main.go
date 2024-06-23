package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/dgrijalva/jwt-go"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("postgres", "user=postgres dbname=payment sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }

    r := mux.NewRouter()
    r.HandleFunc("/auth/register", Register).Methods("POST")
    r.HandleFunc("/auth/login", Login).Methods("POST")

    log.Println("Authentication service started on :8081")
    http.ListenAndServe(":8081", r)
}

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

var jwtKey = []byte("my_secret_key")

func Register(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    _, err = db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", user.Username, string(hashedPassword))
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

func Login(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)

    var storedHash string
    err := db.QueryRow("SELECT password_hash FROM users WHERE username=$1", user.Username).Scan(&storedHash)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
            return
        }
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(user.Password))
    if err != nil {
        http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
        return
    }

    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        Username: user.Username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:    "token",
        Value:   tokenString,
        Expires: expirationTime,
    })
}
