package myHttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"prj2/internal"
	"prj2/logic"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// NOTE: Тут все HTTP-обработчики, роутинг, валидация входных данных, маппинг ошибок в HTTP-ответы.
type HTTPHandlers struct {
	enterprise *logic.Enterprise
	metrics    *Metrics
	shutdown   func()
}

func NewHTTPHandlers(Entp *logic.Enterprise) *HTTPHandlers {
	return &HTTPHandlers{
		enterprise: Entp,
		metrics:    NewMetrics(Entp),
	}
}

func (h *HTTPHandlers) RegisterRoutes(r *mux.Router) {
	r.Use(h.metrics.HTTPMiddleware)
	r.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	r.Handle("/metrics", h.metrics.Handler()).Methods(http.MethodGet)

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
	r.HandleFunc("/enterprise/start", h.StartEntp).Methods(http.MethodPost)
	r.HandleFunc("/enterprise/summary", h.SummaryEntp).Methods(http.MethodGet)
	r.HandleFunc("/enterprise/status", h.StatusEntp).Methods(http.MethodGet)
	r.HandleFunc("/events", h.Events).Methods(http.MethodGet)
	//    NOTE: - Можно отправить запрос на завершение игры
	r.HandleFunc("/enterprise/shutdown", h.ShutdownEntp).Methods(http.MethodPost)

	//    NOTE: - Можно отравить запрос на завершение работы приложения
	r.HandleFunc("/app/close", h.AppClose).Methods(http.MethodPut)

	r.PathPrefix("/").Handler(webHandler()).Methods(http.MethodGet)
}

func (h *HTTPHandlers) SetAppShutdown(shutdown func()) {
	h.shutdown = shutdown
}

func (h *HTTPHandlers) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
	limit := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		limit = parsedLimit
	}

	active := h.enterprise.ActiveMiners(limit)
	list := make([]MinerDTO, 0, len(active))

	for _, m := range active {
		if class != "" && string(m.Class) != class {
			continue
		}
		list = append(list, MinerDTO{
			ID:            m.ID,
			Class:         string(m.Class),
			Energy:        m.Energy,
			IsWorking:     m.IsWorking,
			CoalPerMining: m.CoalPerMining,
		})
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
	catalog := internal.EquipmentCatalog()
	items := make([]EquipmentItemDTO, 0, len(catalog))
	nextGoal, hasNextGoal := internal.NextEquipmentGoal(s.Equipment)

	for index, item := range catalog {
		purchased := s.Equipment[item.Type]
		isNextGoal := hasNextGoal && nextGoal.Type == item.Type
		canBuyNow := !purchased && isNextGoal && s.Balance >= item.Price

		items = append(items, EquipmentItemDTO{
			Type:        string(item.Type),
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			Order:       index + 1,
			Purchased:   purchased,
			CanBuyNow:   canBuyNow,
			IsNextGoal:  isNextGoal,
		})
	}

	hint := "Копите уголь на следующую цель."
	if hasNextGoal {
		hint = "Следующая цель: " + nextGoal.Title
		if s.Balance >= nextGoal.Price {
			hint = "Можно купить следующую цель: " + nextGoal.Title
		}
	} else {
		hint = "Все цели куплены. Можно завершить предприятие и начать новую игру."
	}

	writeJSON(w, http.StatusOK, EquipmentResponse{
		Balance: s.Balance,
		Items:   items,
		Hint:    hint,
	})
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
		Balance: snapshot.Balance, ActiveMiners: active, HiredStats: hired, Equipment: eq,
	})
}

func (h *HTTPHandlers) SummaryEntp(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.summaryResponse())
}

func (h *HTTPHandlers) summaryResponse() EnterpriseSummaryResponse {
	snapshot := h.enterprise.Summary()

	hired := make(map[string]int, len(snapshot.HiredStats))
	for k, v := range snapshot.HiredStats {
		hired[string(k)] = v
	}

	eq := make(map[string]bool, len(snapshot.Equipment))
	for k, v := range snapshot.Equipment {
		eq[string(k)] = v
	}

	goalProgress, goalTotal := equipmentGoalProgress(snapshot.Equipment)
	nextGoal, hasNextGoal := internal.NextEquipmentGoal(snapshot.Equipment)
	nextGoalTitle := ""
	nextGoalPrice := 0
	if hasNextGoal {
		nextGoalTitle = nextGoal.Title
		nextGoalPrice = nextGoal.Price
	}

	return EnterpriseSummaryResponse{
		Balance:       snapshot.Balance,
		ActiveCount:   snapshot.ActiveCount,
		HiredStats:    hired,
		Equipment:     eq,
		GoalProgress:  goalProgress,
		GoalTotal:     goalTotal,
		GoalComplete:  goalProgress == goalTotal,
		NextGoalTitle: nextGoalTitle,
		NextGoalPrice: nextGoalPrice,
		IsShutdown:    snapshot.IsShutdown,
	}
}

func (h *HTTPHandlers) Events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "streaming is not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	send := func() bool {
		data, err := json.Marshal(h.summaryResponse())
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(w, "event: summary\ndata: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !send() {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if !send() {
				return
			}
		}
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
  - status code: ?
  - response body: JSON ?
*/
func (h *HTTPHandlers) ShutdownEntp(w http.ResponseWriter, r *http.Request) {
	d, err := h.enterprise.Shutdown()
	if err != nil {
		writeLogicErr(w, err)
		return
	}

	b := h.enterprise.Status().Balance
	msg := "The enterprise run is complete. You can restart the game or close the application."
	hiredM, purchasedEq := make(map[string]int, 0), make(map[string]bool, 0)

	snap := h.enterprise.Status()
	for k, v := range snap.HiredStats {
		hiredM[string(k)] = v
	}
	for k, v := range snap.Equipment {
		purchasedEq[string(k)] = v
	}

	writeJSON(w, http.StatusOK, ShutdownResponse{DurationSec: int64(d / time.Second),
		FinalBalance: b, HiredMiners: hiredM, PurchasedEquipment: purchasedEq, Message: msg,
	})
}

func (h *HTTPHandlers) StartEntp(w http.ResponseWriter, r *http.Request) {
	if err := h.enterprise.Start(); err != nil {
		writeLogicErr(w, err)
		return
	}

	h.SummaryEntp(w, r)
}

func (h *HTTPHandlers) AppClose(w http.ResponseWriter, _ *http.Request) {
	_, _ = h.enterprise.Shutdown()
	writeJSON(w, http.StatusOK, "stack shutdown started")

	if h.shutdown != nil {
		go func() {
			time.Sleep(100 * time.Millisecond)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_ = stopComposeSupportContainers(ctx)
			h.shutdown()
		}()
	}
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

func equipmentGoalProgress(equipment map[internal.EquipmentType]bool) (int, int) {
	progress := 0
	goalItems := internal.EquipmentTypes()
	for _, item := range goalItems {
		if equipment[item] {
			progress++
		}
	}

	return progress, len(goalItems)
}
