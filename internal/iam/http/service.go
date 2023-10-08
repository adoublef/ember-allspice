package http

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/adoublef/golang-chi/internal/iam/oauth"
	"github.com/go-chi/chi/v5"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
	// NOTE -- Google currently
	iam *oauth.Authenticator
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService() *Service {
	// iam, err := oauth.NewGoogleAuthenticator()
	iam, err := oauth.NewAuthenticator()
	if err != nil {
		panic(err)
	}
	
	s := &Service{
		m: chi.NewMux(),
		iam: iam,
	}
	s.routes()
	return s
}

func (s *Service) routes() {
	s.m.Get("/", s.handleIndex())
	s.m.Get("/signin", s.handleSignIn())
	s.m.Get("/callback", s.handleCallback())
	s.m.Get("/signout", s.handleSignOut())
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}