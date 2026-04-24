package myHttp

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type HTTPServer struct{ httpHandlers *HTTPHandlers }

func NewHTTTPServer(h *HTTPHandlers) *HTTPServer { return &HTTPServer{httpHandlers: h} }

func (s *HTTPServer) StartServer() error {
	router := mux.NewRouter()
	// NOTE: Важная деталь, позволяет вынести все хендлеры в отдельную функцию или пакет
	// NOTE: + к этому, не нужно проверять в ручную методы хендлеров
	s.httpHandlers.RegisterRoutes(router)

	if err := http.ListenAndServe(":8080", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) { // не является ошибкой
			return nil
		}

		return err
	}
	return nil
}
