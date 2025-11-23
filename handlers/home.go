package handlers

import (
	"net/http"
	"os"
)

func ServeHome(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}
