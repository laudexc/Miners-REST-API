package logic

import (
	"context"
	"prj2/internal"
	"slices"
	"sync"
	"time"
)

type Enterprise struct {
	mu sync.RWMutex

	balance      int
	activeMiners map[int]*internal.MinerState
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
}

func NewEnterprise() *Enterprise {
	e := &Enterprise{ // создание компании
		activeMiners: make(map[int]*internal.MinerState),
		hiredStats:   make(map[internal.MinerClass]int),
		equipment:    make(map[internal.EquipmentType]bool),
		incomeCh:     make(chan int),
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
	go e.incomeAggregator() // TODO: что это?

	e.wg.Add(1)
	go e.passiveIncomeLoop() // запустить пассивную добычу

	return nil
}

func (e *Enterprise) HireMiner(class internal.MinerClass) (internal.MinerState, error) {
	profile, ok := internal.MinerProfiles()[class]
	if !ok {
		return internal.MinerState{}, ErrUnknownMinerClass
	}

	e.mu.Lock()
	if !e.isStarted {
		e.mu.Unlock()
		return internal.MinerState{}, ErrNotStarted
	}
	if e.isShutdown {
		e.mu.Unlock()
		return internal.MinerState{}, ErrAlreadyStopped
	}
	if e.balance < profile.Cost {
		e.mu.Unlock()
		return internal.MinerState{}, ErrNotEnoughCoal
	}

	e.balance -= profile.Cost
	e.nextID++

	state := &internal.MinerState{
		ID:            e.nextID,
		Class:         class,
		Energy:        profile.Energy,
		IsWorking:     true,
		CoalPerMining: profile.CoalPerMine,
	}
	e.activeMiners[state.ID] = state
	e.hiredStats[class]++
	e.mu.Unlock()

	e.wg.Add(1)
	go e.runMiner(state.ID, profile)

	return *state, nil
}

func (e *Enterprise) BuyEquipment(equipmentType internal.EquipmentType) error {
	price, ok := internal.EquipmentPrices()[equipmentType]
	if !ok {
		return ErrUnknownEquipmentType
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isStarted {
		return ErrNotStarted
	}
	if e.isShutdown {
		return ErrAlreadyStopped
	}
	if e.equipment[equipmentType] {
		return ErrEquipmentBought
	}
	if e.balance < price {
		return ErrNotEnoughCoal
	}

	e.balance -= price
	e.equipment[equipmentType] = true

	return nil
}

func (e *Enterprise) AddCoal(amount int) error {
	if amount <= 0 {
		return nil
	}

	e.mu.RLock()
	if !e.isStarted {
		e.mu.RUnlock()
		return ErrNotStarted
	}
	ctx := e.ctx
	e.mu.RUnlock()

	select {
	case <-ctx.Done():
		return ErrAlreadyStopped
	case e.incomeCh <- amount:
		return nil
	}
}

func (e *Enterprise) Status() internal.EnterpriseSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	active := make([]internal.MinerState, 0, len(e.activeMiners))
	for _, miner := range e.activeMiners {
		active = append(active, *miner)
	}
	slices.SortFunc(active, func(a, b internal.MinerState) int {
		return a.ID - b.ID
	})

	hired := make(map[internal.MinerClass]int, len(e.hiredStats))
	for class, count := range e.hiredStats {
		hired[class] = count
	}

	equipment := make(map[internal.EquipmentType]bool, len(e.equipment))
	for eq, bought := range e.equipment {
		equipment[eq] = bought
	}

	return internal.EnterpriseSnapshot{
		Balance:      e.balance,
		ActiveMiners: active,
		HiredStats:   hired,
		Equipment:    equipment,
	}
}

func (e *Enterprise) Shutdown() (time.Duration, error) {
	e.mu.Lock()
	if !e.isStarted {
		e.mu.Unlock()
		return 0, ErrNotStarted
	}
	if e.isShutdown {
		e.mu.Unlock()
		return 0, ErrAlreadyStopped
	}

	e.isShutdown = true
	cancel := e.cancel
	startedAt := e.startedAt
	e.mu.Unlock()

	cancel()
	e.wg.Wait()

	return time.Since(startedAt), nil
}

func (e *Enterprise) passiveIncomeLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Second) // TODO: что это?
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
		}

		select {
		case <-e.ctx.Done():
			return
		case e.incomeCh <- 1:
		}
	}
}

func (e *Enterprise) incomeAggregator() {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return

		case amount := <-e.incomeCh:
			e.mu.Lock()
			e.balance += amount
			e.mu.Unlock()
		}
	}
}

// запуск нужного майнера по айди
func (e *Enterprise) runMiner(minerID int, profile internal.MinerProfile) {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Duration(profile.IntervalSec) * time.Second) // TODO: что это?
	defer ticker.Stop()

	for mineIdx := 0; mineIdx < profile.Energy; mineIdx++ {
		select {
		case <-e.ctx.Done():
			e.markMinerStopped(minerID)
			return
		case <-ticker.C:
		}

		coalYield := profile.CoalPerMine + profile.ProgressStep*mineIdx

		select {
		case <-e.ctx.Done():
			e.markMinerStopped(minerID)
			return
		case e.incomeCh <- coalYield:
			e.updateMinerAfterMine(minerID, coalYield)
		}
	}

	e.markMinerStopped(minerID)
}

func (e *Enterprise) updateMinerAfterMine(minerID int, coalYield int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	miner, ok := e.activeMiners[minerID]
	if !ok {
		return
	}

	miner.Energy--
	miner.CoalPerMining = coalYield
}

func (e *Enterprise) markMinerStopped(minerID int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	miner, ok := e.activeMiners[minerID]
	if !ok {
		return
	}

	miner.IsWorking = false
	delete(e.activeMiners, minerID)
}
