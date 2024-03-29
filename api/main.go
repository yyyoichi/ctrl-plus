package main

import (
	"embed"
	"errors"
	"fmt"
	"himakiwa/handlers"
	"himakiwa/handlers/middleware"
	"himakiwa/services"
	"himakiwa/services/database"
	"himakiwa/services/email"
	"himakiwa/services/webpush"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

//go:embed all:out
var assets embed.FS

func init() {
	database.Connect()
}

func main() {
	fmt.Println("hello")
	handler()
}

func handler() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	// init services
	useRepository := services.NewRepositoryServices()
	emailServices := email.NewEmailServices()
	webpushServices := webpush.NewWebpushServices()

	api.HandleFunc("/safe", Safe).Methods(http.MethodGet)

	ah := handlers.NewAutenticateHandlers(emailServices, useRepository)
	api.HandleFunc("/signin", ah.SigninHandler).Methods(http.MethodPost)
	api.HandleFunc("/login", ah.LoginHandler).Methods(http.MethodPost)
	api.HandleFunc("/codein", ah.VerificateHandler).Methods(http.MethodPost)
	api.HandleFunc("/recruitments/{recruitmentUUID}", NotFoundHandler).Methods(http.MethodGet)

	me := api.PathPrefix("/me").Subrouter()
	meHandlers := handlers.NewMeHandlers(useRepository)
	me.HandleFunc("/", meHandlers.MeHandler).Methods(http.MethodGet)
	me.HandleFunc("/logout", meHandlers.LogoutHandler).Methods(http.MethodPost)

	publicRecruitHandlers := handlers.NewPublicRecruitHandlers(useRepository)
	me.HandleFunc("/recruitments", publicRecruitHandlers).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
	me.Use(middleware.AuthMiddleware)

	wp := me.PathPrefix("/webpush").Subrouter()
	wp.HandleFunc("/vapid_public_key", handlers.VapidHandler).Methods(http.MethodGet)
	webpushHandlers := handlers.NewWebPushSubscriptionHandlers(useRepository, webpushServices)
	wp.HandleFunc("/subscription", webpushHandlers).Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	ses := me.PathPrefix("/sessions").Subrouter()
	useRepositoryServices := services.NewRepositoryServices()
	ses.HandleFunc("", handlers.NewSessionsHandlers(useRepositoryServices)).Methods(http.MethodGet, http.MethodPost)
	ses.HandleFunc("/{sessionID}", handlers.NewSessionAtHandlers(useRepositoryServices)).Methods(http.MethodGet, http.MethodPut)

	me.HandleFunc("/sessionkey", handlers.NewSessionKeyHandlers(useRepository, webpushServices)).Methods(http.MethodPost)

	chs := me.PathPrefix("/chats").Subrouter()
	chs.HandleFunc("", handlers.NewChatsHandlers(useRepositoryServices)).Methods(http.MethodGet)
	chs.HandleFunc("/{sessionID}", handlers.NewChatsAtHandlers(useRepositoryServices)).Methods(http.MethodGet, http.MethodPost)

	phs := me.PathPrefix("/participants").Subrouter()
	phs.HandleFunc("/{sessionID}", handlers.NewParticipantsAtHandlers(useRepositoryServices)).Methods(http.MethodPost, http.MethodPut)

	api.Use(middleware.CROSMiddleware)
	api.Use(middleware.CSRFMiddleware)

	// route static files
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, r)
}

func Safe(w http.ResponseWriter, r *http.Request) {
	ctoken := csrf.Token(r)
	w.Header().Set("X-CSRF-Token", ctoken)
	w.WriteHeader(http.StatusOK)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := tryRead(r.URL.Path, w)
	if err == nil {
		return
	}
	err = tryRead("404.html", w)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func tryRead(requestedPath string, w http.ResponseWriter) error {
	reqPath := path.Join("out", requestedPath)
	if reqPath == "out" {
		reqPath = "out/index"
	}
	extension := strings.LastIndex(reqPath, ".")
	if extension == -1 {
		reqPath = fmt.Sprintf("%s.html", reqPath)
	}
	fmt.Printf("'GET' to '%s' origin '%s'\n", reqPath, requestedPath)

	// read file
	f, err := assets.Open(reqPath)
	if err != nil {
		log.Printf("'%s' is not found \n", requestedPath)
		return err
	}
	defer f.Close()

	// dir check
	stat, err := f.Stat()
	if err != nil {
		log.Printf("'%s' is found but it cannot get file info \n", requestedPath)
		return err
	}
	if stat.IsDir() {
		return errors.New("path is dir")
	}

	// content type check
	ext := filepath.Ext(requestedPath)
	var contentType string

	if m := mime.TypeByExtension(ext); m != "" {
		contentType = m
	} else {
		contentType = "text/html"
	}

	w.Header().Set("Content-Type", contentType)
	io.Copy(w, f)

	return nil
}
