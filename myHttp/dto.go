package myHttp

type (
	HireMinerRequest struct {
		Class string `json:"class"` // weak|normal|strong
	}

	MinerDTO struct {
		ID            int    `json:"id"`
		Class         string `json:"class"`
		Energy        int    `json:"energy"`
		IsWorking     bool   `json:"is_working"`
		CoalPerMining int    `json:"coal_per_mining"`
	}

	HireMinerResponse struct {
		Miners []MinerDTO `json:"miners"`
	}

	EnterpriseStatusResponse struct {
		Balance       int             `json:"balance"`
		ActiveMiners  []MinerDTO      `json:"active_miners"`
		HiredStats    map[string]int  `json:"hired_stats"`
		Equipment     map[string]bool `json:"equipment"`
		Notifications []string        `json:"notifications"`
	}

	BuyEquipmentResponse struct {
		Type string `json:"type"`
		Ok   bool   `json:"ok"`
	}

	EquipmentItemDTO struct {
		Type      string `json:"type"`
		Title     string `json:"title"`
		Price     int    `json:"price"`
		Purchased bool   `json:"purchased"`
		CanBuyNow bool   `json:"can_buy_now"`
	}

	EquipmentResponse struct {
		Balance int                `json:"balance"`
		Items   []EquipmentItemDTO `json:"items"`
		Hint    string             `json:"hint"`
	}

	ShutdownResponse struct {
		DurationSec        int64           `json:"duration_sec"`
		FinalBalance       int             `json:"final_balance"`
		HiredMiners        map[string]int  `json:"hired_miners"`
		PurchasedEquipment map[string]bool `json:"purchased_equipment"`
		Message            string          `json:"message"`
	}

	ErrorResponse struct {
		Error string `json:"error"`
	}
)
