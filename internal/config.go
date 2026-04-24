package internal

// NOTE: Тут константы и настройки проекта (цены, параметры шахтёров, адрес сервера)

type (
	MinerClass    string
	MinersCount   int
	EquipmentType string
)

const (
	WeakClass   MinerClass = "weak"
	NormalClass MinerClass = "normal"
	StrongClass MinerClass = "strong"

	EquipmentPickaxe     EquipmentType = "pickaxe"
	EquipmentVentilation EquipmentType = "ventilation"
	EquipmentWagon       EquipmentType = "wagon"

	PickaxePrice     = 3000
	VentilationPrice = 15000
	WagonPrice       = 50000
)

type MinerProfile struct { // профиль для каждого майнера
	Cost         int
	Energy       int
	CoalPerMine  int
	IntervalSec  int
	ProgressStep int
}

func MinerProfiles() map[MinerClass]MinerProfile {
	return map[MinerClass]MinerProfile{
		WeakClass: {
			Cost:        5,
			Energy:      30,
			CoalPerMine: 1,
			IntervalSec: 3,
		},

		NormalClass: {
			Cost:        50,
			Energy:      45,
			CoalPerMine: 3,
			IntervalSec: 2,
		},

		StrongClass: {
			Cost:         450,
			Energy:       60,
			CoalPerMine:  10,
			IntervalSec:  1,
			ProgressStep: 3,
		},
	}
}

func EquipmentPrices() map[EquipmentType]int {
	return map[EquipmentType]int{
		EquipmentPickaxe:     PickaxePrice,
		EquipmentVentilation: VentilationPrice,
		EquipmentWagon:       WagonPrice,
	}
}
