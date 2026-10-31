// Horizon BNPL - Domain Models

package main

import "time"

type CreditPlan struct {
	PlanID       string    `json:"plan_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	ProviderID   string    `json:"provider_id"`
	MinAmount    int64     `json:"min_amount"`
	MaxAmount    int64     `json:"max_amount"`
	Installments int       `json:"installments"`
	InterestRate float64   `json:"interest_rate"`
	GraceDays    int       `json:"grace_days"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Status       string    `json:"status"`
	Requirements []string  `json:"requirements"`
	CreatedAt    time.Time `json:"created_at"`
}

type Provider struct {
	ProviderID   string    `json:"provider_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	ContactEmail string    `json:"contact_email"`
	ContactPhone string    `json:"contact_phone"`
	CreditLimit  int64     `json:"credit_limit"`
	UsedAmount   int64     `json:"used_amount"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Organization struct {
	OrgID        string    `json:"org_id"`
	Name         string    `json:"name"`
	NationalID   string    `json:"national_id"`
	CreditLimit  int64     `json:"credit_limit"`
	PlanIDs      []string  `json:"plan_ids"`
	EmployeeCap  int       `json:"employee_cap"`
	ContactEmail string    `json:"contact_email"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Personnel struct {
	PersonnelID string    `json:"personnel_id"`
	OrgID       string    `json:"org_id"`
	NationalID  string    `json:"national_id"`
	FullName    string    `json:"full_name"`
	Mobile      string    `json:"mobile"`
	CreditScore int       `json:"credit_score"`
	CreditLimit int64     `json:"credit_limit"`
	UsedAmount  int64     `json:"used_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Merchant struct {
	MerchantID string    `json:"merchant_id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	TerminalID string    `json:"terminal_id"`
	IBAN       string    `json:"iban"`
	PlanIDs    []string  `json:"plan_ids"`
	FeePercent float64   `json:"fee_percent"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreditTx struct {
	TxID          string    `json:"tx_id"`
	PlanID        string    `json:"plan_id"`
	ProviderID    string    `json:"provider_id"`
	OrgID         string    `json:"org_id,omitempty"`
	PersonnelID   string    `json:"personnel_id"`
	MerchantID    string    `json:"merchant_id"`
	Amount        int64     `json:"amount"`
	Installments  int       `json:"installments"`
	InstallAmount int64     `json:"install_amount"`
	InterestRate  float64   `json:"interest_rate"`
	Status        string    `json:"status"`
	Nonce         string    `json:"nonce"`
	CreatedAt     time.Time `json:"created_at"`
	SettledAt     time.Time `json:"settled_at,omitempty"`
	BlockHash     string    `json:"block_hash,omitempty"`
	MerkleProof   []string  `json:"merkle_proof,omitempty"`
	Signature     string    `json:"signature,omitempty"`
	PubKey        string    `json:"pub_key,omitempty"`
	Address       string    `json:"address,omitempty"`
}

type Installment struct {
	InstallmentID string    `json:"installment_id"`
	TxID          string    `json:"tx_id"`
	Seq           int       `json:"seq"`
	Amount        int64     `json:"amount"`
	DueDate       time.Time `json:"due_date"`
	PaidAt        time.Time `json:"paid_at,omitempty"`
	Status        string    `json:"status"`
}

type LOSApplication struct {
	AppID        string    `json:"app_id"`
	PersonnelID  string    `json:"personnel_id"`
	PlanID       string    `json:"plan_id"`
	Amount       int64     `json:"amount"`
	Installments int       `json:"installments"`
	Status       string    `json:"status"`
	ScoreResult  int       `json:"score_result,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	SubmittedAt  time.Time `json:"submitted_at"`
	DecidedAt    time.Time `json:"decided_at,omitempty"`
}

type AccountingEntry struct {
	EntryID   string    `json:"entry_id"`
	TxID      string    `json:"tx_id"`
	AccountID string    `json:"account_id"`
	Debit     int64     `json:"debit"`
	Credit    int64     `json:"credit"`
	EntryType string    `json:"entry_type"`
	CreatedAt time.Time `json:"created_at"`
}

type Rule struct {
	RuleID     string    `json:"rule_id"`
	Name       string    `json:"name"`
	Scope      string    `json:"scope"`
	ScopeID    string    `json:"scope_id"`
	Conditions []string  `json:"conditions"`
	Actions    []string  `json:"actions"`
	Priority   int       `json:"priority"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditLog struct {
	ID         int64     `json:"id"`
	EventType  string    `json:"event_type"`
	ActorID    string    `json:"actor_id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Payload    string    `json:"payload"`
	PrevHash   string    `json:"prev_hash"`
	CurrHash   string    `json:"curr_hash"`
	CreatedAt  time.Time `json:"created_at"`
}
