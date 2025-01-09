package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/deasdania/integrate-hydra-keto/handlers"
	"github.com/deasdania/integrate-hydra-keto/integrations/hydra"
)

// URL
const (
	BASEURL_HYDRA_ADMIN = "http://localhost:4445/"
	BASEURL_HYDRA       = "http://localhost:4444/"
	BASEURL_KETO        = "http://localhost:4466/"

	// hydra client
	CLIENT_NAME   = "user-clients"
	CLIENT_SECRET = "user-clients-secret"
	CLIENT_ID     = "05c1c1fc-1212-4e09-90f0-acbeb93eb757" // what ever id you get after create the client
)

func main() {
	// Initialize router
	r := mux.NewRouter()
	var logger *logrus.Entry
	hydraIntegrations := hydra.New(logger, BASEURL_HYDRA_ADMIN, BASEURL_HYDRA, CLIENT_NAME, CLIENT_ID)
	hd := handlers.NewHandlers(logger, hydraIntegrations)
	// Routes
	r.HandleFunc("/login", hd.Login).Methods("POST")
	r.HandleFunc("/verify", hd.VerifyTokenHandler).Methods("GET")
	r.HandleFunc("/logout", hd.Logout).Methods("POST") // only for e.g
	r.HandleFunc("/callback", hd.Callback)

	// Start server
	log.Println("Server starting at :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
