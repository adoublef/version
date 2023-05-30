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
	// s.m.Get("<resource>", <version_handler>(map[string|number]http.Handler{"1": <handler>, "2": <handler>})))
	// I can use the "github.com/hashicorp/go-version" package to check the version
	s.m.Get("/", s.handleResource())
}

func (s *Service) handleResource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// header versioning
		// w.Header().Set("Accept", "application/vnd.api.v1+json")
		s.m.Respond(w, r, &map[string]any{
			"version": "v2",
			"message": "hello world",
		}, http.StatusOK)
	}
}

var defaultMux = &muxHandler{chi.NewMux()}

type muxHandler struct {
	// Include versioning here
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
