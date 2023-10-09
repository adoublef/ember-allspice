package http

import (
	"net/http"
)

func (s *Service) handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, err := s.iam.Callback(w, r)
		if err != nil {
			http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// NOTE -- do something with the user info and sessionId

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
