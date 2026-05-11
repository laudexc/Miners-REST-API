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

	EquipmentHeadlamp          EquipmentType = "headlamp"
	EquipmentGloves            EquipmentType = "gloves"
	EquipmentBoots             EquipmentType = "boots"
	EquipmentHelmet            EquipmentType = "helmet"
	EquipmentPickaxe           EquipmentType = "pickaxe"
	EquipmentRespirator        EquipmentType = "respirator"
	EquipmentFirstAidKit       EquipmentType = "first_aid_kit"
	EquipmentGasDetector       EquipmentType = "gas_detector"
	EquipmentRailTracks        EquipmentType = "rail_tracks"
	EquipmentVentilation       EquipmentType = "ventilation"
	EquipmentWagon             EquipmentType = "wagon"
	EquipmentHydraulicDrill    EquipmentType = "hydraulic_drill"
	EquipmentLift              EquipmentType = "lift"
	EquipmentCommandRoom       EquipmentType = "command_room"
	EquipmentRescueStation     EquipmentType = "rescue_station"
	EquipmentProcessingLine    EquipmentType = "processing_line"
	EquipmentAutomatedMine     EquipmentType = "automated_mine"
	EquipmentIndustrialComplex EquipmentType = "industrial_complex"

	HeadlampPrice          = 100
	GlovesPrice            = 250
	BootsPrice             = 500
	HelmetPrice            = 900
	PickaxePrice           = 1500
	RespiratorPrice        = 3000
	FirstAidKitPrice       = 5000
	GasDetectorPrice       = 8000
	RailTracksPrice        = 12000
	VentilationPrice       = 20000
	WagonPrice             = 50000
	HydraulicDrillPrice    = 150000
	LiftPrice              = 500000
	CommandRoomPrice       = 2000000
	RescueStationPrice     = 5000000
	ProcessingLinePrice    = 12000000
	AutomatedMinePrice     = 25000000
	IndustrialComplexPrice = 50000000
)

type EquipmentProfile struct {
	Type        EquipmentType
	Title       string
	Description string
	Price       int
}

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
	prices := make(map[EquipmentType]int, len(EquipmentCatalog()))
	for _, item := range EquipmentCatalog() {
		prices[item.Type] = item.Price
	}

	return prices
}

func EquipmentCatalog() []EquipmentProfile {
	return []EquipmentProfile{
		{Type: EquipmentHeadlamp, Title: "Налобный фонарь", Description: "Первый нормальный свет для спуска под землю.", Price: HeadlampPrice},
		{Type: EquipmentGloves, Title: "Защитные перчатки", Description: "Меньше травм при работе с породой и инструментом.", Price: GlovesPrice},
		{Type: EquipmentBoots, Title: "Шахтерские ботинки", Description: "Устойчивость на мокрых проходах и деревянных настилах.", Price: BootsPrice},
		{Type: EquipmentHelmet, Title: "Усиленная каска", Description: "Базовая защита перед глубокими сменами.", Price: HelmetPrice},
		{Type: EquipmentPickaxe, Title: "Усиленная кирка", Description: "Первый серьезный инструмент для ручной добычи.", Price: PickaxePrice},
		{Type: EquipmentRespirator, Title: "Респиратор", Description: "Защита от угольной пыли и тяжелого воздуха.", Price: RespiratorPrice},
		{Type: EquipmentFirstAidKit, Title: "Аптечка смены", Description: "Запас перевязки, антисептиков и аварийных средств.", Price: FirstAidKitPrice},
		{Type: EquipmentGasDetector, Title: "Датчик газа", Description: "Раннее предупреждение о метане и опасных смесях.", Price: GasDetectorPrice},
		{Type: EquipmentRailTracks, Title: "Рельсовый путь", Description: "Подготовка маршрута для вывоза угля из забоя.", Price: RailTracksPrice},
		{Type: EquipmentVentilation, Title: "Вентиляция", Description: "Постоянный приток воздуха для длинных смен.", Price: VentilationPrice},
		{Type: EquipmentWagon, Title: "Вагонетка", Description: "Больше угля за один рейс из шахты.", Price: WagonPrice},
		{Type: EquipmentHydraulicDrill, Title: "Гидравлический бур", Description: "Переход от ручной добычи к механизированной.", Price: HydraulicDrillPrice},
		{Type: EquipmentLift, Title: "Грузовой подъемник", Description: "Быстрый подъем угля и оборудования на поверхность.", Price: LiftPrice},
		{Type: EquipmentCommandRoom, Title: "Диспетчерская", Description: "Центр контроля смен, оборудования и рисков.", Price: CommandRoomPrice},
		{Type: EquipmentRescueStation, Title: "Спасательная станция", Description: "Резерв безопасности для аварийных ситуаций.", Price: RescueStationPrice},
		{Type: EquipmentProcessingLine, Title: "Линия сортировки угля", Description: "Промышленная подготовка угля после подъема.", Price: ProcessingLinePrice},
		{Type: EquipmentAutomatedMine, Title: "Автоматизированный забой", Description: "Система добычи с минимальным ручным трудом.", Price: AutomatedMinePrice},
		{Type: EquipmentIndustrialComplex, Title: "Шахтный промышленный комплекс", Description: "Финальная цель: полноценный комплекс за 50 млн угля.", Price: IndustrialComplexPrice},
	}
}

func EquipmentTypes() []EquipmentType {
	catalog := EquipmentCatalog()
	types := make([]EquipmentType, 0, len(catalog))
	for _, item := range catalog {
		types = append(types, item.Type)
	}

	return types
}

func EquipmentProfileByType(equipmentType EquipmentType) (EquipmentProfile, bool) {
	for _, item := range EquipmentCatalog() {
		if item.Type == equipmentType {
			return item, true
		}
	}

	return EquipmentProfile{}, false
}

func NextEquipmentGoal(equipment map[EquipmentType]bool) (EquipmentProfile, bool) {
	for _, item := range EquipmentCatalog() {
		if !equipment[item.Type] {
			return item, true
		}
	}

	return EquipmentProfile{}, false
}
