package http

import "net/http"

func (s *Service) handleSignOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.iam.SignOut(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to url, use home for now
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
