package myHttp

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	httpHandlers *HTTPHandlers
}

func NewHTTTPServer(httpHadler *HTTPHandlers) *HTTPServer {
	return &HTTPServer{
		httpHandlers: httpHadler,
	}
}

func (s *HTTPServer) StartServer() error {
	router := mux.NewRouter()

	// - Шахтёры:
	//    TODO: - Можно получить информацию о требуемом размере оплаты труда для каждого класса шахтёров
	router.Path("/miners/prices").Methods("GET").HandlerFunc(//)
	//    TODO: - Можно нанять нового
	router.Path("/miners").Methods("POST").HandlerFunc(//)
	//    TODO: - Можно получить список всех работающих в данный момент
	router.Path("/miners/active").Methods("GET").HandlerFunc(//)
	//    TODO: - Можно получить список всех работающих в данный момент, отфильтровав по классу
	router.Path("/miners/active/{class}").Methods("GET").Queries("class", "{class}").HandlerFunc(//)

	// - Оборудование:
	//    TODO: - Можно получить информацию о стоимости всех видов оборудования
	router.Path("/equipment/prices").Methods("GET").Queries("class", "{class}").HandlerFunc(//)
	//    TODO: - Можно купить новое оборудование
	router.Path("/equipment/{title}/buy").Methods("POST").HandlerFunc(//)
	//    TODO: - Можно получать информацию о том, какое оборудование уже приобретено, а какое — нет
	router.Path("/equipment").Methods("GET").HandlerFunc(//)

	// - Предприятие:
	//    TODO: - Можно получить промежуточную информацию (текущий баланс, сколько каких шахтёров было нанято за всё время, и тд, по желанию)
	router.Path("/enterprise/status").Methods("GET").HandlerFunc(//)
	//    TODO: - Можно отправить запрос на завершение игры
	router.Path("/enterprise/shutdown").Methods("POST").HandlerFunc(//)

	if err := http.ListenAndServe(":8080", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) { // не является ошибкой
			return nil
		}

		return err
	}

	return nil
}
