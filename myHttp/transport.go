package myHttp

import (
	"encoding/json"
	"net/http"
	"os"
	"prj2/internal"
	"prj2/logic"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// NOTE: Тут все HTTP-обработчики, роутинг, валидация входных данных, маппинг ошибок в HTTP-ответы.
type HTTPHandlers struct{ enterprise *logic.Enterprise }

func NewHTTPHandlers(Entp *logic.Enterprise) *HTTPHandlers { return &HTTPHandlers{enterprise: Entp} }

func (h *HTTPHandlers) RegisterRoutes(r *mux.Router) {
	// - Шахтёры:
	//    NOTE: - Можно получить информацию о требуемом размере оплаты труда для каждого класса шахтёров
	r.HandleFunc("/miners/prices", h.MinersSalary).Methods(http.MethodGet)
	//    NOTE: - Можно нанять нового по query параметрам ?class=class&count=count
	r.HandleFunc("/miners/hire", h.HireMiner).Queries("class", "{class}", "count", "{count}").Methods(http.MethodPost)
	//    NOTE: - Можно получить список всех работающих в данный момент + фильтр по классу
	r.HandleFunc("/miners/active", h.ListOfActive).Methods(http.MethodGet)

	// - Оборудование:
	//    NOTE: - Можно получить информацию о стоимости всех видов оборудования
	r.HandleFunc("/equipment/prices", h.EquipmentPrice).Methods(http.MethodGet)
	//    NOTE: - Можно купить новое оборудование
	r.HandleFunc("/equipment/{type}/buy", h.BuyEquipment).Methods(http.MethodPost)
	//    NOTE: - Можно получать информацию о том, какое оборудование уже приобретено, а какое — нет
	r.HandleFunc("/equipment", h.PurchasedEquipment).Methods(http.MethodGet)

	// - Предприятие:
	//    NOTE: - Можно получить промежуточную информацию (текущий баланс, сколько каких шахтёров было нанято за всё время, и тд, по желанию)
	r.HandleFunc("/enterprise/status", h.StatusEntp).Methods(http.MethodGet)
	//    NOTE: - Можно отправить запрос на завершение игры
	r.HandleFunc("/enterprise/shutdown", h.ShutdownEntp).Methods(http.MethodPost)

	//    NOTE: - Можно отравить запрос на завершение работы приложения
	r.HandleFunc("/app/close", h.AppClose).Methods(http.MethodPut)
}

/*
pattern: /miners/prices
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON with miners classes and hiring prices

failed:
  - status code: ?
  - response body: ?
*/
func (h *HTTPHandlers) MinersSalary(w http.ResponseWriter, r *http.Request) {
	profiles := internal.MinerProfiles()
	writeJSON(w, http.StatusOK, profiles)
}

/*
pattern: /miners
method:  POST
info:    Query params

succeed:
  - status code: 201 Created
  - response body: JSON represent created miner

failed:
  - status code: ?
  - response body: ?
*/
func (h *HTTPHandlers) HireMiner(w http.ResponseWriter, r *http.Request) {
	class := r.URL.Query().Get("class")
	class = strings.TrimSpace(class)
	countStr := r.URL.Query().Get("count")

	count, err := strconv.Atoi(countStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if class == "" {
		writeErr(w, http.StatusBadRequest, "the class field cannot be empty")
		return
	}
	if int(count) < 1 { // если передано количество майнеров для найма < 1
		writeErr(w, http.StatusBadRequest, logic.ErrMinersWrongQuantity.Error())
		return
	}

	miners, err := h.enterprise.HireMiner(internal.MinerClass(class), internal.MinersCount(count))
	if err != nil {
		writeLogicErr(w, err)
		return
	}

	respMiners := make([]MinerDTO, 0, len(miners))
	for _, m := range miners {
		respMiners = append(respMiners, MinerDTO{
			ID:            m.ID,
			Class:         string(m.Class),
			Energy:        m.Energy,
			IsWorking:     m.IsWorking,
			CoalPerMining: m.CoalPerMining,
		})
	}
	writeJSON(w, http.StatusCreated, HireMinerResponse{Miners: respMiners})
}

/*
pattern: /miners/active
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON represent all active miners

failed:
  - status code: ?
  - response body: JSON ?
*/
func (h *HTTPHandlers) ListOfActive(w http.ResponseWriter, r *http.Request) {
	class := r.URL.Query().Get("class")
	active := h.enterprise.Status().ActiveMiners
	list := make([]internal.MinerState, 0, len(active))

	for _, m := range active {
		if class != "" && string(m.Class) != class {
			continue
		}
		list = append(list, m)
	}

	writeJSON(w, http.StatusOK, list)
}

/*
pattern: /equipment/prices
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON with equipment prices

failed:
  - status code: ?
  - response body: JSON ?
*/
func (h *HTTPHandlers) EquipmentPrice(w http.ResponseWriter, r *http.Request) {
	eqPrices := internal.EquipmentPrices()
	writeJSON(w, http.StatusOK, eqPrices)
}

/*
pattern: /equipment/{title}/buy
method:  POST
info:    equipment title in URL path

succeed:
  - status code: 200 OK
  - response body: JSON represent purchase result

failed:
  - status code: ?
  - response body: JSON ?
*/
func (h *HTTPHandlers) BuyEquipment(w http.ResponseWriter, r *http.Request) {
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
  - status code: ?
  - response body: JSON ?
*/
func (h *HTTPHandlers) PurchasedEquipment(w http.ResponseWriter, r *http.Request) {
	s := h.enterprise.Status()
	purchased := s.Equipment
	writeJSON(w, http.StatusOK, purchased)
}

/*
pattern: /enterprise/status
method:  GET
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON represent current enterprise snapshot

failed:
  - status code: ?
  - response body: JSON ?
*/
func (h *HTTPHandlers) StatusEntp(w http.ResponseWriter, r *http.Request) {
	snapshot := h.enterprise.Status()
	active := make([]MinerDTO, 0, len(snapshot.ActiveMiners))

	for _, m := range snapshot.ActiveMiners {
		active = append(active, MinerDTO{
			ID: m.ID, Class: string(m.Class), Energy: m.Energy,
			IsWorking: m.IsWorking, CoalPerMining: m.CoalPerMining,
		})
	}

	hired := make(map[string]int, len(snapshot.HiredStats))
	for k, v := range snapshot.HiredStats {
		hired[string(k)] = v
	}

	eq := make(map[string]bool, len(snapshot.Equipment))
	for k, v := range snapshot.Equipment {
		eq[string(k)] = v
	}

	writeJSON(w, http.StatusOK, EnterpriseStatusResponse{
		Balance: snapshot.Balance, ActiveMiners: active, HiredStats: hired, Equipment: eq, Notifications: snapshot.Notifications,
	})
}

/*
pattern: /enterprise/shutdown
method:  POST
info:    no input required

succeed:
  - status code: 200 OK
  - response body: JSON represent shutdown result + game duration

failed:
  - status code: ?
  - response body: JSON ?
*/
func (h *HTTPHandlers) ShutdownEntp(w http.ResponseWriter, r *http.Request) {
	d, err := h.enterprise.Shutdown()
	if err != nil {
		writeLogicErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ShutdownResponse{DurationSec: int64(d / time.Second)})
}

func (h *HTTPHandlers) AppClose(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, "goodbye.")

	go func() {
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()
}

// Быстрая отправка логической ошибки клиенту
func writeLogicErr(w http.ResponseWriter, err error) {
	writeErr(w, http.StatusBadRequest, err.Error())
}

// Быстрая отправка ошибки клиенту
func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, ErrorResponse{
		Error: msg,
	})
}

// Быстрая отправка JSON ответа с сериализацией для клиента
func writeJSON(w http.ResponseWriter, code int, v any) { // v - это мапа/структура, которую надо сериализовать для JSON
	w.Header().Set("Content-Type", "application/json") // говорим клиенту явно: в ответе JSON
	w.WriteHeader(code)                                // поставить переданный HTTP статус для заголовка
	_ = json.NewEncoder(w).Encode(v)                   // v (структуру/мапу), сериализует в JSON и пишет прямо в ответ
}
