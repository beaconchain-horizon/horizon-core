// Horizon BNPL - Loan Origination System + Rules Engine
// Handles application scoring, rule evaluation, approval flow

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// LOSEngine - موتور پردازش درخواست‌های اعتبار
type LOSEngine struct {
	plans       map[string]*CreditPlan
	providers   map[string]*Provider
	orgs        map[string]*Organization
	personnel   map[string]*Personnel
	apps        map[string]*LOSApplication
	rules       []*Rule
}

func NewLOSEngine() *LOSEngine {
	return &LOSEngine{
		plans:     make(map[string]*CreditPlan),
		providers: make(map[string]*Provider),
		orgs:      make(map[string]*Organization),
		personnel: make(map[string]*Personnel),
		apps:      make(map[string]*LOSApplication),
		rules:     []*Rule{},
	}
}

// ScoringInput - ورودی ارزیابی
type ScoringInput struct {
	PersonnelID string
	PlanID      string
	Amount      int64
	Installments int
}

// ScoringResult - نتیجه ارزیابی
type ScoringResult struct {
	Score    int
	Approved bool
	Reasons  []string
	Limit    int64
}

// Score - محاسبه امتیاز اعتباری بر اساس قوانین
func (e *LOSEngine) Score(in ScoringInput) (*ScoringResult, error) {
	p, ok := e.personnel[in.PersonnelID]
	if !ok {
		return nil, errors.New("personnel not found")
	}
	plan, ok := e.plans[in.PlanID]
	if !ok {
		return nil, errors.New("plan not found")
	}

	result := &ScoringResult{Approved: true, Reasons: []string{}}

	// Amount range check
	if in.Amount < plan.MinAmount {
		result.Approved = false
		result.Reasons = append(result.Reasons, "amount below plan minimum")
	}
	if in.Amount > plan.MaxAmount {
		result.Approved = false
		result.Reasons = append(result.Reasons, "amount above plan maximum")
	}

	// Available credit
	available := p.CreditLimit - p.UsedAmount
	if in.Amount > available {
		result.Approved = false
		result.Reasons = append(result.Reasons, "insufficient credit limit")
	}

	// Plan active window
	now := time.Now()
	if now.Before(plan.StartDate) || now.After(plan.EndDate) {
		result.Approved = false
		result.Reasons = append(result.Reasons, "plan not in active window")
	}

	// Score = simple formula (can be extended)
	result.Score = p.CreditScore

	// Apply rules engine
	ruleResult := e.evaluateRules(in, p, plan)
	if !ruleResult.passed {
		result.Approved = false
		result.Reasons = append(result.Reasons, ruleResult.reasons...)
	}

	result.Limit = available
	return result, nil
}

type ruleOutcome struct {
	passed  bool
	reasons []string
}

func (e *LOSEngine) evaluateRules(in ScoringInput, p *Personnel, plan *CreditPlan) ruleOutcome {
	out := ruleOutcome{passed: true}
	for _, r := range e.rules {
		if !r.Enabled {
			continue
		}
		// Simple condition evaluation (extend as needed)
		for _, c := range r.Conditions {
			ok := evalCondition(c, in, p, plan)
			if !ok {
				out.passed = false
				out.reasons = append(out.reasons, fmt.Sprintf("rule %s: %s failed", r.RuleID, c))
			}
		}
	}
	return out
}

func evalCondition(cond string, in ScoringInput, p *Personnel, plan *CreditPlan) bool {
	// Placeholder - implement actual condition parser
	_ = in
	_ = p
	_ = plan
	return true
}

// Submit - ثبت درخواست جدید
func (e *LOSEngine) Submit(req *LOSApplication) error {
	sc, err := e.Score(ScoringInput{
		PersonnelID:  req.PersonnelID,
		PlanID:       req.PlanID,
		Amount:       req.Amount,
		Installments: req.Installments,
	})
	if err != nil {
		return err
	}
	req.ScoreResult = sc.Score
	req.SubmittedAt = time.Now()

	if sc.Approved {
		req.Status = "approved"
		req.DecidedAt = time.Now()
	} else {
		req.Status = "rejected"
		req.DecidedAt = time.Now()
		req.Notes = fmt.Sprintf("reasons: %v", sc.Reasons)
	}

	req.AppID = newID("app")
	e.apps[req.AppID] = req
	return nil
}

func newID(prefix string) string {
	b := make([]byte, 8)
	rand.Read(b)
	h := sha256.Sum256(b)
	return prefix + "_" + hex.EncodeToString(h[:8])
}

func (e *LOSEngine) GetApp(id string) (*LOSApplication, bool) {
	a, ok := e.apps[id]
	return a, ok
}

func (e *LOSEngine) GetPlan(id string) (*CreditPlan, bool) {
	p, ok := e.plans[id]
	return p, ok
}

func (e *LOSEngine) AddPlan(p *CreditPlan) {
	e.plans[p.PlanID] = p
}

func (e *LOSEngine) AddPersonnel(p *Personnel) {
	e.personnel[p.PersonnelID] = p
}

func (e *LOSEngine) AddProvider(p *Provider) {
	e.providers[p.ProviderID] = p
}

func (e *LOSEngine) AddOrg(o *Organization) {
	e.orgs[o.OrgID] = o
}

func (e *LOSEngine) AddRule(r *Rule) {
	e.rules = append(e.rules, r)
}
