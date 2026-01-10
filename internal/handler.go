package internal

import (
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
	tmpl, err := template.ParseFiles(filepath.Join("templates", "partials", "hello_world.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{"Message": message})
}

func (h *Handler) HelloUKHandler(w http.ResponseWriter, r *http.Request) {
	message := h.greeter.Greet(LocationUK)
	tmpl, err := template.ParseFiles(filepath.Join("templates", "partials", "hello_uk.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]string{"Message": message})
}
