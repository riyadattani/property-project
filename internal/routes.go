package internal

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", handler.IndexHandler).Methods("GET")
	r.HandleFunc("/hello-world", handler.HelloWorldHandler).Methods("GET")
	r.HandleFunc("/hello-uk", handler.HelloUKHandler).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return r
}
