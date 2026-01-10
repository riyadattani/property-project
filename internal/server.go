package internal

import (
	"fmt"
	"net/http"
)

type Server struct {
	addr   string
	router http.Handler
}

func NewServer(cfg Config) *http.Server {

	greeter := NewGreeter()
	handler := NewHandler(greeter)
	router := NewRouter(handler)

	addr := fmt.Sprintf(":%s", cfg.Port)

	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}
