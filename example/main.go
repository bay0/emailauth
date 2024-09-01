package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bay0/emailauth"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort       string
	RedisAddr        string
	SMTPHost         string
	SMTPPort         string
	SMTPUsername     string
	SMTPPassword     string
	SenderEmail      string
	AllowUnencrypted bool
}

type RedisClientWrapper struct {
	*redis.Client
}

func (r *RedisClientWrapper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClientWrapper) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisClientWrapper) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

type Server struct {
	authService *emailauth.AuthService
	router      *mux.Router
}

func NewServer(authService *emailauth.AuthService) *Server {
	s := &Server{
		authService: authService,
		router:      mux.NewRouter(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/send-code", s.handleSendCode).Methods("POST")
	s.router.HandleFunc("/verify-code", s.handleVerifyCode).Methods("POST")
}

func (s *Server) handleSendCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.authService.SendAuthCode(ctx, req.Email); err != nil {
		log.Printf("Error sending auth code: %v", err)
		http.Error(w, fmt.Sprintf("Failed to send auth code: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"message": "Auth code sent successfully"}, http.StatusOK)
}

func (s *Server) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	isValid, err := s.authService.VerifyCode(ctx, req.Email, req.Code)
	if err != nil {
		log.Printf("Error verifying code: %v", err)
		http.Error(w, fmt.Sprintf("Failed to verify code: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"valid": isValid}, http.StatusOK)
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func loadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	allowUnencrypted, _ := strconv.ParseBool(os.Getenv("ALLOW_UNENCRYPTED"))

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		SMTPHost:         os.Getenv("SMTP_HOST"),
		SMTPPort:         os.Getenv("SMTP_PORT"),
		SMTPUsername:     os.Getenv("SMTP_USERNAME"),
		SMTPPassword:     os.Getenv("SMTP_PASSWORD"),
		SenderEmail:      os.Getenv("SENDER_EMAIL"),
		AllowUnencrypted: allowUnencrypted,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})
	defer redisClient.Close()

	emailSender := emailauth.NewSMTPEmailSender(
		config.SMTPHost,
		config.SMTPPort,
		config.SMTPUsername,
		config.SMTPPassword,
		config.SenderEmail,
		config.AllowUnencrypted,
	)

	codeStore := emailauth.NewRedisCodeStore(&RedisClientWrapper{redisClient})

	authService := emailauth.NewAuthService(emailSender, codeStore)

	server := NewServer(authService)

	log.Printf("Server starting on :%s", config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, server.router))
}
