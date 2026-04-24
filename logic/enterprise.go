package logic

import (
	"context"
	"fmt"
	"prj2/internal"
	"slices"
	"sync"
	"time"
)

type Enterprise struct {
	mu sync.RWMutex

	balance      int
	activeMiners map[int]*internal.MinerState    // все работающие майнеры
	hiredStats   map[internal.MinerClass]int     // статы всех нанятых майнеров
	equipment    map[internal.EquipmentType]bool // купленное оборудование

	incomeCh chan int // канал в который идет прибыль

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	startedAt  time.Time
	isStarted  bool
	isShutdown bool
	nextID     int

	notifyStep        int
	nextNotifyBalance int
	notifications     []string
}

func NewEnterprise() *Enterprise {
	e := &Enterprise{ // создание компании
		activeMiners: make(map[int]*internal.MinerState),
		hiredStats:   make(map[internal.MinerClass]int),
		equipment:    make(map[internal.EquipmentType]bool),
		incomeCh:     make(chan int),

		notifyStep:        250,
		nextNotifyBalance: 250,
		notifications:     make([]string, 0),
	}

	for equipmentType := range internal.EquipmentPrices() {
		// все купленное оборудование инициализируется как false
		e.equipment[equipmentType] = false
	}

	return e
}

func (e *Enterprise) Start() error { // старт базовой компании с добычей 1 уголь/c
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isStarted {
		return ErrAlreadyStarted
	}

	e.ctx, e.cancel = context.WithCancel(context.Background()) // добавить контекст с отменой
	e.startedAt = time.Now()
	e.isStarted = true

	e.wg.Add(1)
	go e.incomeAggregator() // incomeAggregator постоянно читает incomeCh в который кидают уголь и прибавляет в balance

	e.wg.Add(1)
	go e.passiveIncomeLoop() // запустить пассивную добычу

	return nil
}

func (e *Enterprise) HireMiner(class internal.MinerClass, count internal.MinersCount) ([]*internal.MinerState, error) { // нанять майнера
	profile, ok := internal.MinerProfiles()[class] // internal.MinerProfiles() возвращает map[MinerClass]MinerProfile
	// Дальше [class] берёт профиль по ключу (классу шахтёра)
	if !ok {
		return nil, ErrUnknownMinerClass
	}

	e.mu.Lock()
	if int(count) < 1 { // если передано количество майнеров для найма < 1
		e.mu.Unlock()
		return nil, ErrMinersWrongQuantity
	}
	if !e.isStarted { // если майнер не запущен
		e.mu.Unlock()
		return nil, ErrNotStarted
	}
	if e.isShutdown { // если все завершено
		e.mu.Unlock()
		return nil, ErrAlreadyStopped
	}
	if e.balance < profile.Cost*int(count) { // если баланс меньше стоимости шахтера
		e.mu.Unlock()
		return nil, ErrNotEnoughCoal
	}

	miners := make([]*internal.MinerState, 0, int(count))
	for i := 0; i < int(count); i++ {
		e.balance -= profile.Cost // успешная покупка выбранного майнера
		e.nextID++                // дается следующий айди

		state := &internal.MinerState{ // генерация статов купленного майнера
			ID:            e.nextID,
			Class:         class,
			Energy:        profile.Energy,
			IsWorking:     true,
			CoalPerMining: profile.CoalPerMine,
		}
		miners = append(miners, state)

		e.activeMiners[state.ID] = state // добавить нового шахтера в активных майнеров
		e.hiredStats[class]++            // увеличить счетчик сколько всего нанято шахтеров
	}
	e.mu.Unlock()

	for _, m := range miners {
		e.wg.Add(1)
		go e.runMiner(m.ID, profile) // запустить в горутине майнера с выбранными параметрами
	}

	return miners, nil // вернуть данные нанятого шахтёра
}

func (e *Enterprise) BuyEquipment(equipmentType internal.EquipmentType) error { // покупка оборудования
	price, ok := internal.EquipmentPrices()[equipmentType] // EquipmentPrices() возращает мапу и дальше через ключ [equipmentType] поиск цены
	if !ok {
		return ErrUnknownEquipmentType
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isStarted { // если не работает
		return ErrNotStarted
	}
	if e.isShutdown { // если завершена работа
		return ErrAlreadyStopped
	}
	if e.equipment[equipmentType] { // если оборудование уже куплено
		return ErrEquipmentBought
	}
	if e.balance < price { // если не хватает денег на балансе для покупки
		return ErrNotEnoughCoal
	}

	e.balance -= price                // списание денег
	e.equipment[equipmentType] = true // метка, что такое то оборудование куплено

	return nil
}

func (e *Enterprise) AddCoal(amount int) error { // положить уголь в канал
	if amount <= 0 { // если ничего не пришло, или пришло некорректное количество угля
		return nil
	}

	e.mu.RLock()
	if !e.isStarted { // если не работает предприятие
		e.mu.RUnlock()
		return ErrNotStarted
	}
	ctx := e.ctx // получить снимок контекста из структуры, чтобы не напороться на гонку данных
	e.mu.RUnlock()

	select {
	case <-ctx.Done():
		return ErrAlreadyStopped
	case e.incomeCh <- amount: // положить в канал уголь
		return nil
	}
}

func (e *Enterprise) Status() internal.EnterpriseSnapshot { // статус предприятия
	e.mu.RLock()
	defer e.mu.RUnlock()

	active := make([]internal.MinerState, 0, len(e.activeMiners)) // сделать снимок количества активных майнеров
	for _, miner := range e.activeMiners {                        // пройтись по активным майнерам
		active = append(active, *miner) // добавить майнеров в новый слайс
	}
	slices.SortFunc(active, func(a, b internal.MinerState) int { // отсортировать их по айди
		return a.ID - b.ID
	})

	hired := make(map[internal.MinerClass]int, len(e.hiredStats)) // сделать снимок сколько всего нанято майнеров
	for class, count := range e.hiredStats {                      // пройтись по всем работникам
		hired[class] = count // скопировать их классы и количество в новую мапу
	}

	equipment := make(map[internal.EquipmentType]bool, len(e.equipment)) // сделать снимок сколько и какое оборудование куплено
	for eq, bought := range e.equipment {                                // пройтись по всему оборудованию на предприятии и добавить в новую мапу
		equipment[eq] = bought
	}

	notifications := append([]string(nil), e.notifications...) // все последние уведомления

	return internal.EnterpriseSnapshot{ // вернуть снимок запрашиваемых данных
		Balance:       e.balance,
		ActiveMiners:  active,
		HiredStats:    hired,
		Equipment:     equipment,
		Notifications: notifications,
	}
}

func (e *Enterprise) Shutdown() (time.Duration, error) { // завершение всех процессов
	e.mu.Lock()
	if !e.isStarted { // если ничего и не было запущено, то ошибка
		e.mu.Unlock()
		return 0, ErrNotStarted
	}
	if e.isShutdown { // если уже завершено, завершать нечего
		e.mu.Unlock()
		return 0, ErrAlreadyStopped
	}

	e.isShutdown = true      // метка что завершена работа
	cancel := e.cancel       // просто копия функции отмены в локальную переменную
	startedAt := e.startedAt // копия метки времени когда стартовало
	e.mu.Unlock()

	cancel()    // отмена контекста
	e.wg.Wait() // подождать завершения всех горутин

	return time.Since(startedAt), nil // сколько предприятие проработало
}

func (e *Enterprise) passiveIncomeLoop() { // постоянная добыча +1 без условий
	defer e.wg.Done()

	ticker := time.NewTicker(time.Second * 2) // таймер, тикающий каждые 2 секунды пока идет добыча
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C: // подожди, пока тикер отправит следующий тик, должны пройти 2 сек
		}

		select {
		case <-e.ctx.Done():
			return
		case e.incomeCh <- 1: // положить в канал 1 уголь
		}
	}
}

func (e *Enterprise) incomeAggregator() { // постоянно читает канал incomeCh и добавляет уголь в balance
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done(): // если контекст завершен
			return

		case amount := <-e.incomeCh: // берём число угля из общего канала дохода
			e.mu.Lock()
			e.balance += amount // добавляем в баланс это число
			for e.balance >= e.nextNotifyBalance {
				str := fmt.Sprintf("На балансе есть как минимум %d угля", e.nextNotifyBalance)
				fmt.Println(str)
				e.nextNotifyBalance += e.notifyStep // уведомление и том сколько есть угля

				e.notifications = append(e.notifications, str)
				if len(e.notifications) > 8 {
					e.notifications = e.notifications[len(e.notifications)-8:] // оставить самые новые записи
				}
			}
			e.mu.Unlock()
		}
	}
}

// запуск нужного майнера по айди
func (e *Enterprise) runMiner(minerID int, profile internal.MinerProfile) {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Duration(profile.IntervalSec) * time.Second) // таймер шахтёра:
	// как часто он делает одну добычу - 1/2/3 сек в зависимости от класса
	defer ticker.Stop()

	for mineIdx := 0; mineIdx < profile.Energy; mineIdx++ { // начать копать на сколько хватит энергии
		select {
		case <-e.ctx.Done(): // если контекст завершился, то майнер останавливается
			e.markMinerStopped(minerID)
			return
		case <-ticker.C:
		}

		coalYield := profile.CoalPerMine + profile.ProgressStep*mineIdx // добытый уголь = уголь за добычу + прогресс за каждый добытый уголь(у слабого и обычного = 0)

		select {
		case <-e.ctx.Done():
			e.markMinerStopped(minerID) // проверка завершения контекста
			return
		case e.incomeCh <- coalYield: // положить добытый уголь в канал
			e.updateMinerAfterMine(minerID, coalYield) // обновить сведения майнера
		}
	}

	e.markMinerStopped(minerID) // когда энергия заканчивается - майнер останавливается
}

func (e *Enterprise) updateMinerAfterMine(minerID int, coalYield int) { // обновить сведения о майнере по айди
	e.mu.Lock()
	defer e.mu.Unlock()

	miner, ok := e.activeMiners[minerID] // проверить по списку всех майнеров, найти нужного по айди
	if !ok {
		return
	}

	miner.Energy--                  // -1 энергия за копание
	miner.CoalPerMining = coalYield // это обновление “текущей мощности добычи” шахтёра в состоянии.
	// Нужно для того, у которого добыча растёт с каждым циклом.
}

func (e *Enterprise) markMinerStopped(minerID int) { // остановить раскопки майнера
	e.mu.Lock()
	defer e.mu.Unlock()

	miner, ok := e.activeMiners[minerID] // проверить по списку всех майнеров, найти нужного по айди
	if !ok {
		return
	}

	miner.IsWorking = false         // поставить метку что майнер не работает
	delete(e.activeMiners, minerID) // удалить майнера из активных
}
