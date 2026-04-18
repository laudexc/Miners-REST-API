package internal

type MinerState struct {
	ID            int
	Class         MinerClass
	Energy        int
	IsWorking     bool
	CoalPerMining int
}

type EnterpriseSnapshot struct {
	Balance      int
	ActiveMiners []MinerState
	HiredStats   map[MinerClass]int
	Equipment    map[EquipmentType]bool
}
