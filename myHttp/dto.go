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
		Balance      int             `json:"balance"`
		ActiveMiners []MinerDTO      `json:"active_miners"`
		HiredStats   map[string]int  `json:"hired_stats"`
		Equipment    map[string]bool `json:"equipment"`
	}

	EnterpriseSummaryResponse struct {
		Balance       int             `json:"balance"`
		ActiveCount   int             `json:"active_count"`
		HiredStats    map[string]int  `json:"hired_stats"`
		Equipment     map[string]bool `json:"equipment"`
		GoalProgress  int             `json:"goal_progress"`
		GoalTotal     int             `json:"goal_total"`
		GoalComplete  bool            `json:"goal_complete"`
		NextGoalTitle string          `json:"next_goal_title"`
		NextGoalPrice int             `json:"next_goal_price"`
		IsShutdown    bool            `json:"is_shutdown"`
	}

	BuyEquipmentResponse struct {
		Type string `json:"type"`
		Ok   bool   `json:"ok"`
	}

	EquipmentItemDTO struct {
		Type        string `json:"type"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Price       int    `json:"price"`
		Order       int    `json:"order"`
		Purchased   bool   `json:"purchased"`
		CanBuyNow   bool   `json:"can_buy_now"`
		IsNextGoal  bool   `json:"is_next_goal"`
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
