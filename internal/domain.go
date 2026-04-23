package internal

type MinerState struct { // структура майнера
	ID            int
	Class         MinerClass
	Energy        int
	IsWorking     bool
	CoalPerMining int
}

type EnterpriseSnapshot struct { // структура для снимка всего что имеется на предприятии
	Balance      int
	ActiveMiners []MinerState
	HiredStats   map[MinerClass]int
	Equipment    map[EquipmentType]bool
	Notifications []string
}
