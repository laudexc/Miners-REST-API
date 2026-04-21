package myHttp

import (
	"encoding/json"
	"net/http"
	"prj2/internal"
	"prj2/logic"
	"time"

	"github.com/gorilla/mux"
)

// internal/transport/http
// HTTP-обработчики, роутинг, валидация входных данных, маппинг ошибок в HTTP-ответы.
type HTTPHandlers struct{ enterprise *logic.Enterprise }

func NewHTTPHandlers(Entp *logic.Enterprise) *HTTPHandlers { return &HTTPHandlers{enterprise: Entp} }

func (h *HTTPHandlers) RegisterRoutes(r *mux.Router) {
	// - Шахтёры:
	//    TODO: - Можно получить информацию о требуемом размере оплаты труда для каждого класса шахтёров
	r.HandleFunc("/miners/salary", h.MinersSalary).Methods(http.MethodGet)
	//    TODO: - Можно нанять нового
	r.HandleFunc("/miners", h.HireMiner).Methods(http.MethodPost)
	//    TODO: - Можно получить список всех работающих в данный момент
	r.HandleFunc("/miners/active", h.ListOfActive).Methods(http.MethodGet)
	//    TODO: - Можно получить список всех работающих в данный момент, отфильтровав по классу
	r.HandleFunc("/miners/active/{class}", h.ListActiveByClass).Methods(http.MethodGet).Queries("class", "{class}")

	// - Оборудование:
	//    TODO: - Можно получить информацию о стоимости всех видов оборудования
	r.HandleFunc("/equipment/prices", h.EquipmentPrice).Methods(http.MethodGet).Queries("class", "{class}")
	//    TODO: - Можно купить новое оборудование
	r.HandleFunc("/equipment/{title}/buy", h.BuyEquipment).Methods(http.MethodPost)
	//    TODO: - Можно получать информацию о том, какое оборудование уже приобретено, а какое — нет
	r.HandleFunc("/equipment", h.PurchasedEquipment).Methods(http.MethodGet)

	// - Предприятие:
	//    TODO: - Можно получить промежуточную информацию (текущий баланс, сколько каких шахтёров было нанято за всё время, и тд, по желанию)
	r.HandleFunc("/enterprise/status", h.TakeSnapshotEntp).Methods(http.MethodGet)
	//    TODO: - Можно отправить запрос на завершение игры
	r.HandleFunc("/enterprise/shutdown", h.ShutdownGame).Methods(http.MethodPost)
}

func writeLogicErr(w http.ResponseWriter, err error) {
	writeErr(w, http.StatusBadRequest, err.Error())
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, ErrorResponse{
		Error: msg,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) { // v - это мапа/структура, которую надо сериализовать для JSON
	w.Header().Set("Content-Type", "application/json") // говорим клиенту явно: в ответе JSON
	w.WriteHeader(code)                                // поставить переданный HTTP статус для заголовка
	_ = json.NewEncoder(w).Encode(v)                   // v (структуру/мапу), сериализует в JSON и пишет прямо в ответ
}

/*
pattern: /miners/prices
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON with miners classes and hiring prices

failed:
  - status code: 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) MinersSalary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be GET!"})
		return
	}
}

/*
pattern: /miners
method:  POST
info:    JSON in HTTP request body (miner class)

succeed:
  - status code: 201 Created
  - response body: JSON represent created miner

failed:
  - status code: 400, 404, 409, 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) HireMiner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be POST!"})
		return
	}

}

/*
pattern: /miners/active
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON represent all active miners

failed:
  - status code: 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) ListOfActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be GET!"})
		return
	}

}

/*
pattern: /miners/active/{class}?class={class}
method:  GET
info:    class in URL path/query

succeed:
  - status code: 200 OK
  - response body: JSON represent active miners filtered by class

failed:
  - status code: 400, 404, 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) ListActiveByClass(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be GET!"})
		return
	}

}

/*
pattern: /equipment/prices
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON with equipment prices

failed:
  - status code: 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) EquipmentPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be GET!"})
		return
	}

}

/*
pattern: /equipment/{title}/buy
method:  POST
info:    equipment title in URL path

succeed:
  - status code: 200 OK
  - response body: JSON represent purchase result

failed:
  - status code: 400, 404, 409, 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) BuyEquipment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be POST!"})
		return
	}

	t := mux.Vars(r)["type"]
	if err := h.enterprise.BuyEquipment(internal.EquipmentType(t)); err != nil {
		writeLogicErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuyEquipmentResponse{Type: t, Ok: true})
}

/*
pattern: /equipment
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON represent purchased/not purchased equipment

failed:
  - status code: 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) PurchasedEquipment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be GET!"})
		return
	}
}

/*
pattern: /enterprise/status
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON represent current enterprise snapshot

failed:
  - status code: 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) TakeSnapshotEntp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be GET!"})
		return
	}

}

/*
pattern: /enterprise/shutdown
method:  POST
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON represent shutdown result + game duration

failed:
  - status code: 400, 409, 500...
  - response body: JSON with error + time
*/
func (h *HTTPHandlers) ShutdownGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "method should be POST!"})
		return
	}

	d, err := h.enterprise.Shutdown()
	if err != nil {
		writeLogicErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ShutdownResponse{DurationSec: int64(d / time.Second)})
}
