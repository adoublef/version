package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *muxHandler
}

// ServeHTTP implements http.Handler
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService() *Service {
	s := Service{defaultMux}

	s.routes()
	return &s
}

func (s *Service) routes() {
	s.m.Get("/", s.handleResource())
}

func (s *Service) handleResource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.m.Respond(w, r, &map[string]any{
			"firstname": "John",
			"surname": "Brown",
		}, http.StatusOK)
	}
}

var defaultMux = &muxHandler{chi.NewMux()}

type muxHandler struct {
	chi.Router
}

func (m *muxHandler) Respond(w http.ResponseWriter, r *http.Request, data any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, "could not encode in json", code)
		}
	}
}
