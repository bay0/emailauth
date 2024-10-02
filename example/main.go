package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bay0/emailauth"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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
	SessionKey       string
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
	store       *sessions.CookieStore
	templates   *template.Template
}

func NewServer(authService *emailauth.AuthService, sessionKey string) *Server {
	s := &Server{
		authService: authService,
		router:      mux.NewRouter(),
		store:       sessions.NewCookieStore([]byte(sessionKey)),
		templates:   template.Must(template.ParseGlob("templates/*.html")),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.handleHome).Methods("GET")
	s.router.HandleFunc("/login", s.handleLoginForm).Methods("GET")
	s.router.HandleFunc("/login", s.handleLoginSubmit).Methods("POST")
	s.router.HandleFunc("/verify", s.handleVerifyForm).Methods("GET")
	s.router.HandleFunc("/verify", s.handleVerifySubmit).Methods("POST")
	s.router.HandleFunc("/logout", s.handleLogout).Methods("GET")

	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func (s *Server) renderTemplate(w http.ResponseWriter, tmpl string, data map[string]interface{}) {
	log.Printf("Rendering template: %s", tmpl)
	if data == nil {
		data = make(map[string]interface{})
	}
	data["Title"] = "Email Authentication"

	t, err := template.Must(s.templates.Clone()).ParseFiles("templates/layout.html", "templates/"+tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, "session")
	email, ok := session.Values["email"].(string)
	authenticated, _ := session.Values["authenticated"].(bool)
	if !ok || !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	s.renderTemplate(w, "home.html", map[string]interface{}{
		"Email": email,
	})
}
func (s *Server) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, "session")
	flash, _ := session.Values["flash"].(string)
	delete(session.Values, "flash")
	session.Save(r, w)
	s.renderTemplate(w, "login.html", map[string]interface{}{
		"Flash": flash,
	})
}

func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if email == "" {
		s.renderTemplate(w, "login.html", map[string]interface{}{
			"Flash":     "Email is required",
			"FlashType": "error",
		})
		return
	}

	ctx := r.Context()
	if err := s.authService.SendAuthCode(ctx, email); err != nil {
		log.Printf("Error sending auth code: %v", err)
		s.renderTemplate(w, "login.html", map[string]interface{}{
			"Flash":     "Failed to send auth code",
			"FlashType": "error",
		})
		return
	}

	session, _ := s.store.Get(r, "session")
	session.Values["email"] = email
	session.Save(r, w)

	http.Redirect(w, r, "/verify", http.StatusSeeOther)
}

func (s *Server) handleVerifyForm(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, "session")
	email, ok := session.Values["email"].(string)
	if !ok || email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	flash, _ := session.Values["flash"].(string)
	delete(session.Values, "flash")
	session.Save(r, w)
	s.renderTemplate(w, "verify.html", map[string]interface{}{
		"Email": email,
		"Flash": flash,
	})
}

func (s *Server) handleVerifySubmit(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, "session")
	email, ok := session.Values["email"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		s.renderTemplate(w, "verify.html", map[string]interface{}{
			"Email":     email,
			"Flash":     "Code is required",
			"FlashType": "error",
		})
		return
	}

	ctx := r.Context()
	isValid, err := s.authService.VerifyCode(ctx, email, code)
	if err != nil {
		log.Printf("Error verifying code: %v", err)
		s.renderTemplate(w, "verify.html", map[string]interface{}{
			"Email":     email,
			"Flash":     "Failed to verify code",
			"FlashType": "error",
		})
		return
	}

	if !isValid {
		s.renderTemplate(w, "verify.html", map[string]interface{}{
			"Email":     email,
			"Flash":     "Invalid code",
			"FlashType": "error",
		})
		return
	}

	session.Values["authenticated"] = true
	session.Values["flash"] = "Successfully logged in!"
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, "session")
	session.Values["authenticated"] = false
	delete(session.Values, "email")
	session.Values["flash"] = "You have been logged out."
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
		SessionKey:       getEnv("SESSION_KEY", "your-secret-key"),
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

	server := NewServer(authService, config.SessionKey)

	log.Printf("Server starting on :%s", config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, server.router))
}
