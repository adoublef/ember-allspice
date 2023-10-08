package oauth

// https://auth0.com/docs/quickstart/webapp/golang/01-login
import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

// Google
func NewAuthenticator() (*Authenticator, error) {
	p, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("error finding endpoint: %w", err)
	}

	cfg := oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.ExpandEnv("${HOSTNAME}/callback"),
		Endpoint:     p.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	a := Authenticator{Provider: p, Config: cfg}
	return &a, err
}

func (a *Authenticator) SignIn(w http.ResponseWriter, r *http.Request) (string, error) {
	state, err := randString(16)
	if err != nil {
		return "", err
	}

	c := &http.Cookie{
		Name:     "state",
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)

	return a.AuthCodeURL(state), nil
}

func (a *Authenticator) Callback(w http.ResponseWriter, r *http.Request) (*oauth2.Token,*oidc.UserInfo, error){
	state, err := r.Cookie("state")
	if err != nil {
		return nil,nil, fmt.Errorf("state not found: %w", err)
	}
	if r.URL.Query().Get("state") != state.Value {
		return nil,nil, fmt.Errorf("state did not match")
	}
	// Exchange an authorization code for a token.
	token, err := a.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		return nil,nil, fmt.Errorf("failed to verify id token: %w", err)
	}

	u, err := a.UserInfo(r.Context(), oauth2.StaticTokenSource(token))
	if err != nil {
		return nil,nil, fmt.Errorf("failed to get userinfo: %w", err)
	}

	// TODO -- create session and sessionID

	return token,u, nil
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
