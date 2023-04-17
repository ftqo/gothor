package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/ftqo/gothor/config"
	"github.com/ftqo/gothor/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type Server struct {
	DB       db.Querier
	Log      zerolog.Logger
	Server   http.Server
	Sessions *scs.SessionManager
}

func (s *Server) Start(c config.Server, wg *sync.WaitGroup) error {
	defer wg.Done()

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(&apiLogger{s.Log}))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)

	s.Server = http.Server{
		Addr:    ":" + strconv.Itoa(int(c.Port)),
		Handler: s.Sessions.LoadAndSave(Handler(s, WithRouter(r), WithServerBaseURL(c.BaseURL), WithMiddlewares(s.middlewares()))),
	}

	s.Log.Info().Msgf("starting server on %s", s.Server.Addr)
	err := s.Server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting server: %v", err)
	}
	return nil
}

func (s *Server) Stop() {
	deadline, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	_ = s.Server.Shutdown(deadline)
}

func (s *Server) middlewares() map[string]func(http.Handler) http.Handler {
	return map[string]func(http.Handler) http.Handler{
		"authentication": s.Sessions.LoadAndSave,
	}
}
