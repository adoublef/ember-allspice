package main

import (
	"net/http"

	"github.com/adoublef/golang-chi/html"
)

func handleAbout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Title": "About",
		}

		if err := html.Execute(w, "about", data); err != nil {
			http.Error(w, "error writing partial "+err.Error(), http.StatusInternalServerError)
		}
	}
}
