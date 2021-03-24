// Package models provides gorm models and related structs
package models

import (
	"encoding/json"
	"time"
)

const (
	dbDateFormat = "2006-01-02"
)

// PayPeriod struct (model)
type PayPeriod struct {
	StartDate time.Time `gorm:"primaryKey" json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// Custom Marshal for PayPeriod.
// Converts time.Time to formatted string
func (p *PayPeriod) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}{
		StartDate: p.StartDate.Format(dbDateFormat),
		EndDate:   p.EndDate.Format(dbDateFormat),
	})
}
