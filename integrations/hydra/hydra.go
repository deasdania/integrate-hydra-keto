package hydra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	client "github.com/ory/hydra-client-go/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"golang.org/x/oauth2"
)

type Hydra struct {
	logger        *logrus.Entry
	hydraAdminURL string
	hydraURL      string
	clientName    string
	clientID      string
}

type HydraInterface interface {
	CreateClient(ctx context.Context, req *client.OAuth2Client) error
	IntrospectToken(token, requiredScope string) (*TokenIntrospect, error)
	AuthorizationCheck(ctx context.Context, requiredScope string) error
	RequestToken(ctx context.Context, username, password string) (*oauth2.Token, error)
}

const (
	introspectPath    = "/admin/oauth2/introspect"
	generateTokenPath = "/oauth2/token"
)

// New creates a new Auth instance.
func New(logger *logrus.Entry, hydraAdminURL, hydraURL, clientName, clientID string) *Hydra {
	return &Hydra{
		logger:        logger,
		hydraAdminURL: hydraAdminURL,
		hydraURL:      hydraURL,
		clientName:    clientName,
		clientID:      clientID,
	}
}

// to create new client
func (h *Hydra) CreateClient(ctx context.Context, req *client.OAuth2Client) error {
	oAuth2Client := *client.NewOAuth2Client() // OAuth2Client |
	oAuth2Client.SetClientId(*req.ClientId)
	oAuth2Client.SetClientName(*req.ClientName)
	oAuth2Client.SetClientSecret(*req.ClientSecret)
	oAuth2Client.SetRegistrationClientUri(*req.RegistrationClientUri)
	oAuth2Client.SetGrantTypes(req.GrantTypes)
	oAuth2Client.SetScope(*req.Scope)

	configuration := client.NewConfiguration()
	configuration.Servers = []client.ServerConfiguration{
		{
			URL: h.hydraAdminURL, // Public API URL
		},
	}
	apiClient := client.NewAPIClient(configuration)
	resp, r, err := apiClient.OAuth2API.CreateOAuth2Client(context.Background()).OAuth2Client(oAuth2Client).Execute()
	if err != nil {
		switch r.StatusCode {
		case http.StatusConflict:
			h.logger.Infof("Conflict when creating oAuth2Client: %v\n", err)
		default:
			h.logger.Infof("Error when calling `OAuth2Api.CreateOAuth2Client``: %v\n", err)
			h.logger.Infof("Full HTTP response: %v\n", r)
		}
		return err
	}
	h.logger.Infof("Created client with name %s\n", resp.GetClientName())
	return nil
}

// Function to introspect the token using Hydra
func (h *Hydra) IntrospectToken(token, requiredScope string) (*TokenIntrospect, error) {
	// Create the request to introspect the token
	token = strings.TrimPrefix(token, "Bearer ")
	req, err := http.NewRequest("POST", h.hydraAdminURL+introspectPath, strings.NewReader("token="+token+"&scope="+requiredScope))
	if err != nil {
		h.logger.Errorf("Error creating introspection request: %v\n", err)
		return nil, err
	}

	// Set the required headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Perform the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.logger.Errorf("Error during HTTP request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the response body into TokenIntrospect struct
	var tokenInfo TokenIntrospect
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		h.logger.Errorf("Error decoding response body: %v\n", err)
		return nil, err
	}

	return &tokenInfo, nil
}

// Function to check if the token has the required scope
func hasScope(tokenScope string, requiredScope string) bool {
	scopes := strings.Fields(tokenScope)
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

func (h *Hydra) AuthorizationCheck(ctx context.Context, requiredScope string) error {
	// Get the token from the request header
	// Extract metadata (headers) from the context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("no metadata in context")
	}

	// Extract the token from the Authorization header
	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return fmt.Errorf("unauthorized: missing token")
	}

	// Introspect the token using Hydra
	tokenInfo, err := h.IntrospectToken(tokens[0], requiredScope)
	if err != nil || !tokenInfo.Active {
		h.logger.Errorf("Error during token introspection: %v\n", err)
		return fmt.Errorf("Invalid token")
	}

	// Check if the token has the required scope for accessing the resource
	if !hasScope(tokenInfo.Scope, requiredScope) {
		h.logger.Error("Insufficient scope")
		return fmt.Errorf("Insufficient scope")
	}

	return nil
}

func (h *Hydra) RequestToken(ctx context.Context, username, password string) (*oauth2.Token, error) {
	var oauth2Config = oauth2.Config{
		ClientID:     h.clientID,   // Client ID registered with Hydra
		ClientSecret: h.clientName, // Client Secret registered with Hydra
		Endpoint: oauth2.Endpoint{
			TokenURL: h.hydraURL + generateTokenPath, // Hydra token endpoint
		},
	}
	// Request the token from Hydra's /oauth2/token endpoint
	token, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		log.Fatalf("Error obtaining token: %v", err)
		return nil, err
	}
	return token, nil
}
