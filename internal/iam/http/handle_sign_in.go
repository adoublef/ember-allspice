package http

import (
	"net/http"
)

func (s *Service) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := s.iam.SignIn(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}
