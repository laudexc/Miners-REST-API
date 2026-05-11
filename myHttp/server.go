package myHttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HTTPServer struct{ httpHandlers *HTTPHandlers }

func NewHTTTPServer(h *HTTPHandlers) *HTTPServer { return &HTTPServer{httpHandlers: h} }

func (s *HTTPServer) StartServer() error {
	router := mux.NewRouter()
	s.httpHandlers.RegisterRoutes(router)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	s.httpHandlers.SetAppShutdown(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}
	return nil
}
