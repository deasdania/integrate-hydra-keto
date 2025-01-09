package hydra

type TokenIntrospect struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope"`
	ClientID  string `json:"client_id"`
	Sub       string `json:"sub"`
	Exp       int    `json:"exp"`
	Iat       int    `json:"iat"`
	Nbf       int    `json:"nbf"`
	Aud       []any  `json:"aud"`
	Iss       string `json:"iss"`
	TokenType string `json:"token_type"`
	TokenUse  string `json:"token_use"`
}
