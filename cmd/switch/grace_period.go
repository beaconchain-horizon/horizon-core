package main

import (
	"time"
)

const (
	gracePeriodDays   = 30
	graceDailyTxLimit = 100
)

// GraceStatus describes the current grace-period state
// of an expired license.
type GraceStatus struct {
	InGracePeriod   bool   `json:"in_grace_period"`
	DaysOverdue     int    `json:"days_overdue"`
	DaysLeftInGrace int    `json:"days_left_in_grace"`
	DailyLimit      int    `json:"daily_limit"`
	DailyUsed       int64  `json:"daily_used"`
	DailyRemaining  int64  `json:"daily_remaining"`
	IsReadOnly      bool   `json:"is_read_only"`
	ExpiresAt       int64  `json:"expires_at"`
	Message         string `json:"message"`
}

// getGraceStatus computes the grace status for a license.
//
// Rules:
//   - now <= expires_at          -> active (InGracePeriod=false, IsReadOnly=false)
//   - expires_at < now <= +30d   -> grace (InGracePeriod=true, daily limit)
//   - now > expires_at + 30d     -> read-only (IsReadOnly=true)
func getGraceStatus(license *License) GraceStatus {
	if license == nil {
		return GraceStatus{
			IsReadOnly: true,
			Message:    "no active license",
		}
	}

	now := time.Now()
	expiresAt := time.Unix(license.ExpiresAt, 0)

	// Still active
	if !now.After(expiresAt) {
		return GraceStatus{
			ExpiresAt: license.ExpiresAt,
			Message:   "license is active",
		}
	}

	daysOverdue := int(now.Sub(expiresAt).Hours() / 24)
	daysLeft := gracePeriodDays - daysOverdue

	used := getTodayTxCount()
	remaining := int64(graceDailyTxLimit) - used
	if remaining < 0 {
		remaining = 0
	}

	status := GraceStatus{
		InGracePeriod:   daysOverdue <= gracePeriodDays,
		DaysOverdue:     daysOverdue,
		DaysLeftInGrace: daysLeft,
		DailyLimit:      graceDailyTxLimit,
		DailyUsed:       used,
		DailyRemaining:  remaining,
		ExpiresAt:       license.ExpiresAt,
	}

	if daysOverdue > gracePeriodDays {
		status.IsReadOnly = true
		status.Message = "grace period exceeded: read-only mode"
	} else {
		status.Message = "grace period: limited daily transactions"
	}

	return status
}

// getTodayTxCount returns the number of transactions
// created since midnight local time.
func getTodayTxCount() int64 {
	now := time.Now()
	startOfDay := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location(),
	)

	var count int64
	db.Model(&Transaction{}).
		Where("timestamp >= ?", startOfDay.Unix()).
		Count(&count)

	return count
}

// isTransactionAllowedInGrace returns whether a new
// transaction should be accepted, given the current grace
// status.
//
// Active licenses are always allowed.
// Grace licenses are allowed until the daily limit.
// Read-only licenses are never allowed.
func isTransactionAllowedInGrace() (bool, string) {
	license := getActiveLicenseOrNil()
	if license == nil {
		return false, "no active license"
	}

	status := getGraceStatus(license)

	if status.IsReadOnly {
		return false, "grace period exceeded: read-only mode"
	}

	if status.InGracePeriod && status.DailyRemaining <= 0 {
		return false, "grace period daily limit exceeded"
	}

	return true, ""
}

// getActiveLicenseOrNil returns the current license
// regardless of validity state, or nil if none exists.
func getActiveLicenseOrNil() *License {
	if _, lic, _ := getLicenseState(); lic != nil {
		return lic
	}
	l, _ := getActiveLicense()
	return l
}
