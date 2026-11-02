package db

import (
	"log"
)

type GlobalBank struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	SWIFT   string `json:"swift"`
	Status  string `json:"status"`
}

type GlobalIndustry struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	Sector  string `json:"sector"`
	Country string `json:"country"`
	Status  string `json:"status"`
}

func SeedGlobalEntities() {
	banks := []GlobalBank{
		{Name: "JPMorgan Chase", Country: "USA", SWIFT: "CHASUS33", Status: "Active"},
		{Name: "Bank of America", Country: "USA", SWIFT: "BOFAUS3N", Status: "Active"},
		{Name: "ICBC (Industrial and Commercial Bank of China)", Country: "China", SWIFT: "ICBKCNBJ", Status: "Active"},
		{Name: "Sberbank", Country: "Russia", SWIFT: "SABRRUMM", Status: "Active"},
		{Name: "HSBC", Country: "UK", SWIFT: "HBUKGB4B", Status: "Active"},
		{Name: "Deutsche Bank", Country: "Germany", SWIFT: "DEUTDEFF", Status: "Active"},
		{Name: "BNP Paribas", Country: "France", SWIFT: "BNPAFRPP", Status: "Active"},
		{Name: "Mitsubishi UFJ Financial Group", Country: "Japan", SWIFT: "BOTKJPJT", Status: "Active"},
	}

	industries := []GlobalIndustry{
		{Name: "Siemens", Sector: "SCADA / Industrial Automation", Country: "Germany", Status: "Active"},
		{Name: "General Electric (GE)", Sector: "Energy / Power Grid", Country: "USA", Status: "Active"},
		{Name: "Shell", Sector: "Oil & Gas", Country: "UK", Status: "Active"},
		{Name: "Toyota", Sector: "Manufacturing", Country: "Japan", Status: "Active"},
		{Name: "Saudi Aramco", Sector: "Energy", Country: "Saudi Arabia", Status: "Active"},
		{Name: "Gazprom", Sector: "Energy", Country: "Russia", Status: "Active"},
	}

	for _, bank := range banks { DB.FirstOrCreate(&bank, GlobalBank{Name: bank.Name}) }
	for _, ind := range industries { DB.FirstOrCreate(&ind, GlobalIndustry{Name: ind.Name}) }
	log.Println("🌍 Global Banks and Industries Seeded Successfully.")
}
