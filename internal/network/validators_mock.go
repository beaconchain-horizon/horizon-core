package network

// ValidatorData ساختار داده ولیدیتور برای داشبورد
type ValidatorData struct {
	Index      int     `json:"index"`
	Status     string  `json:"status"`
	Balance    float64 `json:"balance"`
	PublicKey  string  `json:"public_key"`
	TotalIncome float64 `json:"total_income"`
}

// GetMockValidators داده‌های اولیه برای نمایش در داشبورد
func GetMockValidators() []ValidatorData {
	return []ValidatorData{
		{Index: 1, Status: "active", Balance: 32.5, PublicKey: "0x1a2b...", TotalIncome: 12.4},
		{Index: 2, Status: "active", Balance: 32.5, PublicKey: "0x3c4d...", TotalIncome: 18.1},
		{Index: 3, Status: "offline", Balance: 32.5, PublicKey: "0x5e6f...", TotalIncome: 0},
		{Index: 4, Status: "active", Balance: 32.5, PublicKey: "0x7a8b...", TotalIncome: 20.3},
		{Index: 5, Status: "active", Balance: 32.5, PublicKey: "0x9c0d...", TotalIncome: 15.7},
	}
}
