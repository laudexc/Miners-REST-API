package myHttp

// type ErrorDTO struct {
// 	Msg  string
// 	Time time.Time
// }

// func (ed ErrorDTO) ToString() string {
// 	b, err := json.MarshalIndent(ed, "", "    ")
// 	if err != nil {
// 		log.Fatalln("Impossible error in MarshalIndent", err)
// 	}

// 	return string(b)
// }

// func NewErrorDTO(err error) ErrorDTO {
// 	errDTO := ErrorDTO{
// 		Msg:  err.Error(),
// 		Time: time.Now(),
// 	}

// 	return errDTO
// }

type HireMinerRequest struct {
	Class string `json:"class"` // weak|normal|strong
}

type MinerDTO struct {
	ID            int    `json:"id"`
	Class         string `json:"class"`
	Energy        int    `json:"energy"`
	IsWorking     bool   `json:"is_working"`
	CoalPerMining int    `json:"coal_per_mining"`
}

type HireMinerResponse struct {
	Miner MinerDTO `json:"miner"`
}

type EnterpriseStatusResponse struct {
	Balance      int             `json:"balance"`
	ActiveMiners []MinerDTO      `json:"active_miners"`
	HiredStats   map[string]int  `json:"hired_stats"`
	Equipment    map[string]bool `json:"equipment"`
}

type BuyEquipmentResponse struct {
	Type string `json:"type"`
	Ok   bool   `json:"ok"`
}

type ShutdownResponse struct {
	DurationSec int64 `json:"duration_sec"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
