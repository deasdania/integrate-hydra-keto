package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"github.com/deasdania/integrate-hydra-keto/integrations/hydra"
)

type Handlers struct {
	hydraI hydra.HydraInterface
	logger *logrus.Entry
}

func NewHandlers(logger *logrus.Entry, hydraI hydra.HydraInterface) *Handlers {
	return &Handlers{
		logger: logger,
		hydraI: hydraI,
	}
}

// User credentials (In a real app, you'd have a database)
var validUser = map[string]string{
	"username": "password123", // In production, you should hash passwords
}

// Secret key for signing the JWT token
var jwtKey = []byte("my_secret_key")

// Struct for holding the JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Login handler to issue a token
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Check if the user exists and the password matches
	if password != validUser[username] {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// TODO: integrate with hydra
	// c := r.Context()
	// token, err := h.hydraI.RequestToken(c, username, password)
	// if err != nil {
	// 	h.logger.Errorf("got Error: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	return
	// }

	// Create the JWT claims, which contains the username and expiration time
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the token using the claims and signing it with the secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	// Return the token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, tokenString)))
}

func VerifyTokenStr(tokenString string) error {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Return the secret key to validate the token
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token")
	}

	// If the token is valid, we can extract the claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		fmt.Printf("Token valid! Welcome %s\n", claims.Username)

		return nil
	} else {
		return fmt.Errorf("invalid token")
	}
}

// TODO: verify with introspect OAauth2 Access and Refresh Token
// VerifyToken handler to verify if the JWT is valid
func (h *Handlers) VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) == 0 {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	if len(tokenString) > 6 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Verify the token
	err := VerifyTokenStr(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Token is valid"))
}

// Logout handler to "invalidate" the token (not implemented here but should be in a real-world scenario)
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// In a real app, you would invalidate the token by adding it to a blacklist.
	// For simplicity, this just responds with a message saying the user is logged out.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logged out successfully"}`))
}

func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code from the query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Log the code and state to confirm it's working
	log.Printf("Received authorization code: %s\n", code)
	log.Printf("Received state: %s\n", state)

	// Optionally, you could verify the state parameter for security, but we won't process it here.

	// Simply send a response back indicating the code was received.
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><body><h1>Callback Received</h1><p>Authorization Code: %s</p></body></html>", code)
}
