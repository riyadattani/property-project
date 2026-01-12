package internal

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
)

type Handler struct {
	greeter Greeter
}

func NewHandler(greeter Greeter) *Handler {
	return &Handler{greeter: greeter}
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func (h *Handler) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	message := h.greeter.Greet(LocationWorld)
	h.renderGreeting(w, message)
}

func (h *Handler) HelloUKHandler(w http.ResponseWriter, r *http.Request) {
	message := h.greeter.Greet(LocationUK)
	h.renderGreeting(w, message)
}

func (h *Handler) renderGreeting(w http.ResponseWriter, message string) {
	tmpl, err := template.ParseFiles(filepath.Join("templates", "partials", "greeting.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{"Message": message})
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
