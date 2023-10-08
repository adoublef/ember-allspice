package http

import (
	"net/http"

	"github.com/adoublef/golang-chi/html"
)

func (s*Service) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Title": "Welcome",
			"Name":  "Golang",
		}

		if err := html.Execute(w, "index", data); err != nil {
			http.Error(w, "error writing partial "+err.Error(), http.StatusInternalServerError)
		}
	}
}
