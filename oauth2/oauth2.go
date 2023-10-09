package oauth2

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type ConfigOption func(*oauth2.Config)

func WithEndpoint(url oauth2.Endpoint) ConfigOption {
	return func(c *oauth2.Config) {
		c.Endpoint = url
	}
}

func WithRedirect(url string) ConfigOption {
	return func(c *oauth2.Config) {
		c.RedirectURL = url
	}
}

func WithScope(scopes ...string) ConfigOption {
	return func(c *oauth2.Config) {
		// do I need to create this?
		c.Scopes = append(c.Scopes, scopes...)
	}
}

// NewGoogleConfigFromProvider creates an OAuth2 configuration.
//
// Requires environment variables:
//
//	```
//	1. GOOGLE_CLIENT_ID
//	2. GOOGLE_CLIENT_SECRET
//	```
func NewGoogleConfigFromProvider(p *oidc.Provider, opts ...ConfigOption) (oauth2.Config, error) {
	c := oauth2.Config{
		ClientID:     LookUpEnv("GOOGLE_CLIENT_ID"),
		ClientSecret: LookUpEnv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     p.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID},
	}

	for _, o := range opts {
		o(&c)
	}

	// NOTE -- need a better check
	// if ok := (c.Endpoint == oauth2.Endpoint{}); !ok {
	// 	return (oauth2.Config{}), fmt.Errorf("endpoint is required")
	// }

	return c, nil
}

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

// NewAuthenticator will validate oauth requests for a selected provider
//
// Currently only Google is supported, there
func NewAuthenticator(ctx context.Context, opts ...ConfigOption) (*Authenticator, error) {
	p, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("failed to get provider details: %w", err)
	}

	cfg, err := NewGoogleConfigFromProvider(p, opts...)
	if err != nil {
		return nil, err
	}

	return &Authenticator{p, cfg}, nil
}

// SignIn generates an authorization URL for the given oauth2.Config
//
// @link https://github.com/denoland/deno_kv_oauth/blob/main/lib/sign_in.ts
func (a *Authenticator) SignIn(w http.ResponseWriter, r *http.Request) (string, error) {
	state, err := newUUID()
	if err != nil {
		return "", err
	}
	uri := a.AuthCodeURL(state)
	// TODO -- create session Id
	// TODO -- get success url
	setCookie(w, r, cookieName(r, OAuthCookieName), state,
		isSecure(r), maxAge(10*time.Minute))
	return uri, nil
}

// Callback
func (a *Authenticator) Callback(w http.ResponseWriter, r *http.Request) (sessionId string, profile *oidc.UserInfo, err error) {
	state, err := getCookie(r)
	if err != nil {
		return "", nil, err
	}

	// TODO -- delete session from store

	if r.URL.Query().Get("state") != state {
		return "", nil, fmt.Errorf("state did not match")
	}
	token, err := a.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		return "", nil, fmt.Errorf("failed to verify id token: %w", err)
	}

	// TODO -- create session id
	if sessionId, err = newUUID(); err != nil {
		return "", nil, err
	}

	u, err := a.UserInfo(r.Context(), oauth2.StaticTokenSource(token))
	if err != nil {
		return "", nil, fmt.Errorf("failed to get userinfo: %w", err)
	}

	// TODO -- respond with sessionId and (profile)
	setCookie(w, r, cookieName(r, SiteCookieName), sessionId,
		isSecure(r))
	return sessionId, u, err
}

// SignOut removes the Session ID from the cookies
func (a *Authenticator) SignOut(w http.ResponseWriter, r *http.Request) error {
	setCookie(w, r, cookieName(r, SiteCookieName), "",
		isSecure(r), maxAge(-1))
	return nil
}

// getProfile

//

func LookUpEnv(key string) string {
	s, ok := os.LookupEnv(key)
	if err := fmt.Errorf("missing variable %s", key); !ok {
		panic(err)
	}
	return s
}

func newUUID() (string, error) {
	id, err := uuid.NewRandomFromReader(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %w", err)
	}
	return id.String(), nil
}

type cookieOption func(*http.Cookie)

func maxAge(t time.Duration) cookieOption {
	return func(c *http.Cookie) {
		if t < 0 {
			c.MaxAge = math.MinInt
		} else {
			c.MaxAge = int(t.Seconds())
		}
	}
}

func isSecure(r *http.Request) cookieOption {
	return func(c *http.Cookie) {
		c.Secure = r.TLS != nil
	}
}

func setCookie(w http.ResponseWriter, r *http.Request, name, value string, opts ...cookieOption) {
	c := http.Cookie{
		Name:     name,
		Secure:   true,
		Path:     "/",
		HttpOnly: true,
		Value:    value,
		// 90 days
		MaxAge:   int((90 * 24 * time.Hour).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}

	for _, o := range opts {
		o(&c)
	}

	http.SetCookie(w, &c)
}

// getCookie returns cookie value if it exists else returns an error
func getCookie(r *http.Request) (string, error) {
	name := cookieName(r, OAuthCookieName)
	c, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("failed to find value of cookie %s: %w", name, err)
	}
	return c.Value, nil
}

func cookieName(r *http.Request, name string) string {
	if secure := r.TLS != nil; secure {
		name = "__Host-" + name
	}
	return name
}

const (
	OAuthCookieName = "oauth-session"
	SiteCookieName  = "site-session"
)
