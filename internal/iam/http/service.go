package http

import (
	"context"
	"net/http"
	"os"

	"github.com/adoublef/ember-allspice/oauth2"
	"github.com/go-chi/chi/v5"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
	// NOTE -- Google currently
	iam *oauth2.Authenticator
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService() *Service {
	iam, err := oauth2.NewAuthenticator(context.Background(),
		oauth2.WithRedirect(os.ExpandEnv("${HOSTNAME}/callback")),
		oauth2.WithScope("email", "profile"))
	if err != nil {
		panic(err.Error())
	}

	s := &Service{
		m:   chi.NewMux(),
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
