package server

import (
	"net/http"

	"github.com/expki/backend/pixel-protocol/claude"
	"github.com/expki/backend/pixel-protocol/database"
	"github.com/google/uuid"
)

type Server struct {
	db     *database.Database
	claude *claude.Client
}

func New(db *database.Database, claudeClient *claude.Client) *Server {
	return &Server{
		db:     db,
		claude: claudeClient,
	}
}

// PlayerSecret represents the secret authentication structure
type PlayerSecret struct {
	Secret string `json:"_secret"`
}

// extractSecretFromCookie gets the player secret from cookie
func (s *Server) extractSecretFromCookie(r *http.Request) (uuid.UUID, error) {
	if cookie, err := r.Cookie("player_secret"); err == nil {
		return uuid.Parse(cookie.Value)
	}
	return uuid.Nil, &SecretNotFoundError{}
}

// setPlayerSecretCookie sets a secure, long-lasting cookie with the player secret
func (s *Server) setPlayerSecretCookie(r *http.Request, w http.ResponseWriter, secret uuid.UUID) { // Check if the request was made over HTTPS
	isHTTPS := r.TLS != nil ||
		r.Header.Get("X-Forwarded-Proto") == "https" ||
		r.Header.Get("X-Forwarded-Protocol") == "https" ||
		r.URL.Scheme == "https"

	cookie := &http.Cookie{
		Name:     "player_secret",
		Value:    secret.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0, // Session cookie (no expiration)
	}
	http.SetCookie(w, cookie)
}

// SecretNotFoundError represents when no secret is found in body or cookie
type SecretNotFoundError struct{}

func (e *SecretNotFoundError) Error() string {
	return "no player secret found in request body or cookie"
}
