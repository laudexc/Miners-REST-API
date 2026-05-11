package logic

import (
	"context"
	"prj2/internal"
	"sync"
	"time"
)

const (
	MaxActiveMiners = 1_000_000
	MaxHirePreview  = 100
)

type minerCohort struct {
	StartID       int
	Count         int
	Class         internal.MinerClass
	Energy        int
	CoalPerMining int
	MineIndex     int
	NextMineAt    time.Time
	Interval      time.Duration
}

type Enterprise struct {
	mu sync.RWMutex

	balance      int
	activeMiners int
	cohorts      []minerCohort
	hiredStats   map[internal.MinerClass]int
	equipment    map[internal.EquipmentType]bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	startedAt  time.Time
	isStarted  bool
	isShutdown bool
	nextID     int
}

func NewEnterprise() *Enterprise {
	e := &Enterprise{
		hiredStats: make(map[internal.MinerClass]int),
		equipment:  make(map[internal.EquipmentType]bool),
	}
	e.resetStateLocked()

	return e
}

func (e *Enterprise) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isStarted && !e.isShutdown {
		return ErrAlreadyStarted
	}

	e.resetStateLocked()
	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.startedAt = time.Now()
	e.isStarted = true
	e.isShutdown = false

	e.wg.Add(1)
	go e.simulationLoop(e.ctx)

	return nil
}

func (e *Enterprise) HireMiner(class internal.MinerClass, count internal.MinersCount) ([]*internal.MinerState, error) {
	profile, ok := internal.MinerProfiles()[class]
	if !ok {
		return nil, ErrUnknownMinerClass
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if int(count) < 1 {
		return nil, ErrMinersWrongQuantity
	}
	if !e.isStarted {
		return nil, ErrNotStarted
	}
	if e.isShutdown {
		return nil, ErrAlreadyStopped
	}
	if e.balance < profile.Cost*int(count) {
		return nil, ErrNotEnoughCoal
	}
	if e.activeMiners+int(count) > MaxActiveMiners {
		return nil, ErrActiveMinerLimit
	}

	startID := e.nextID + 1
	e.nextID += int(count)
	e.balance -= profile.Cost * int(count)
	e.activeMiners += int(count)
	e.hiredStats[class] += int(count)

	e.cohorts = append(e.cohorts, minerCohort{
		StartID:       startID,
		Count:         int(count),
		Class:         class,
		Energy:        profile.Energy,
		CoalPerMining: profile.CoalPerMine,
		NextMineAt:    time.Now().Add(time.Duration(profile.IntervalSec) * time.Second),
		Interval:      time.Duration(profile.IntervalSec) * time.Second,
	})

	previewCount := min(int(count), MaxHirePreview)
	miners := make([]*internal.MinerState, 0, previewCount)
	for i := 0; i < previewCount; i++ {
		miners = append(miners, &internal.MinerState{
			ID:            startID + i,
			Class:         class,
			Energy:        profile.Energy,
			IsWorking:     true,
			CoalPerMining: profile.CoalPerMine,
		})
	}

	return miners, nil
}

func (e *Enterprise) BuyEquipment(equipmentType internal.EquipmentType) error {
	profile, ok := internal.EquipmentProfileByType(equipmentType)
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
	nextGoal, ok := internal.NextEquipmentGoal(e.equipment)
	if ok && nextGoal.Type != equipmentType {
		return ErrEquipmentLocked
	}
	if e.balance < profile.Price {
		return ErrNotEnoughCoal
	}

	e.balance -= profile.Price
	e.equipment[equipmentType] = true

	return nil
}

func (e *Enterprise) AddCoal(amount int) error {
	if amount <= 0 {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isStarted {
		return ErrNotStarted
	}
	if e.isShutdown {
		return ErrAlreadyStopped
	}

	e.addCoalLocked(amount)
	return nil
}

func (e *Enterprise) Status() internal.EnterpriseSnapshot {
	summary := e.Summary()
	return internal.EnterpriseSnapshot{
		Balance:      summary.Balance,
		ActiveMiners: e.ActiveMiners(0),
		HiredStats:   summary.HiredStats,
		Equipment:    summary.Equipment,
	}
}

func (e *Enterprise) Summary() internal.EnterpriseSummarySnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return internal.EnterpriseSummarySnapshot{
		Balance:     e.balance,
		ActiveCount: e.activeMiners,
		HiredStats:  copyHiredStats(e.hiredStats),
		Equipment:   copyEquipment(e.equipment),
		IsShutdown:  e.isShutdown,
	}
}

func (e *Enterprise) ActiveMiners(limit int) []internal.MinerState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	total := e.activeMiners
	if limit > 0 && total > limit {
		total = limit
	}

	active := make([]internal.MinerState, 0, total)
	for _, cohort := range e.cohorts {
		for i := 0; i < cohort.Count && (limit <= 0 || len(active) < limit); i++ {
			active = append(active, internal.MinerState{
				ID:            cohort.StartID + i,
				Class:         cohort.Class,
				Energy:        cohort.Energy,
				IsWorking:     true,
				CoalPerMining: cohort.CoalPerMining,
			})
		}
		if limit > 0 && len(active) >= limit {
			break
		}
	}

	return active
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

func (e *Enterprise) simulationLoop(ctx context.Context) {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	passiveTicks := 0
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			e.mu.Lock()
			if e.isShutdown {
				e.mu.Unlock()
				continue
			}

			passiveTicks++
			if passiveTicks%2 == 0 {
				e.addCoalLocked(1)
			}
			e.mineCohortsLocked(now)
			e.mu.Unlock()
		}
	}
}

func (e *Enterprise) mineCohortsLocked(now time.Time) {
	active := e.cohorts[:0]
	for _, cohort := range e.cohorts {
		profile := internal.MinerProfiles()[cohort.Class]
		for cohort.Energy > 0 && !now.Before(cohort.NextMineAt) {
			yieldPerMiner := profile.CoalPerMine + profile.ProgressStep*cohort.MineIndex
			cohort.CoalPerMining = yieldPerMiner
			e.addCoalLocked(yieldPerMiner * cohort.Count)
			cohort.Energy--
			cohort.MineIndex++
			cohort.NextMineAt = cohort.NextMineAt.Add(cohort.Interval)
		}

		if cohort.Energy > 0 {
			active = append(active, cohort)
			continue
		}

		e.activeMiners -= cohort.Count
	}

	e.cohorts = active
}

func (e *Enterprise) addCoalLocked(amount int) {
	e.balance += amount
}

func (e *Enterprise) resetStateLocked() {
	e.balance = 0
	e.activeMiners = 0
	e.cohorts = make([]minerCohort, 0)
	e.hiredStats = make(map[internal.MinerClass]int)
	e.equipment = make(map[internal.EquipmentType]bool)
	for _, equipmentType := range internal.EquipmentTypes() {
		e.equipment[equipmentType] = false
	}
	e.nextID = 0
}

func copyHiredStats(src map[internal.MinerClass]int) map[internal.MinerClass]int {
	dst := make(map[internal.MinerClass]int, len(src))
	for class, count := range src {
		dst[class] = count
	}
	return dst
}

func copyEquipment(src map[internal.EquipmentType]bool) map[internal.EquipmentType]bool {
	dst := make(map[internal.EquipmentType]bool, len(src))
	for eq, bought := range src {
		dst[eq] = bought
	}
	return dst
}
