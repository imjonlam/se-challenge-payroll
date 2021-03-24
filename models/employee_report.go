// Package models provides gorm models and related structs
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EmployeeReports struct
type EmployeeReports struct {
	EmployeeReports []EmployeeReport `json:"employeeReports"`
}

// EmployeeReport struct (model)
type EmployeeReport struct {
	EmployeeID  uint      `gorm:"primaryKey" json:"employeeID"`
	PayPeriodID time.Time `gorm:"primaryKey"`
	PayPeriod   PayPeriod `json:"payPeriod"`
	Amount      float32   `json:"amountPaid"`
}

// Custom Marshal for EmployeeReport.
// Removes PayPeriodID and formats Amount
func (e *EmployeeReport) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		EmployeeID uint      `json:"employeeID"`
		PayPeriod  PayPeriod `json:"payPeriod"`
		Amount     string    `json:"amountPaid"`
	}{
		EmployeeID: e.EmployeeID,
		PayPeriod:  e.PayPeriod,
		Amount:     fmt.Sprintf("$%.2f", e.Amount),
	})
}

// Insert a new entry into the employee_reports table.
// Updates existing entry if an existing record is found
func (e *EmployeeReport) CreateEmpoyeeReport(db *gorm.DB, amount float32) error {
	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "EmployeeID, StartDate"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"Amount": gorm.Expr("Amount + ?", amount)}),
	}).Create(e).Error

	if err != nil {
		return err
	}

	return nil
}

// Retrieve all records in employee_reports table
func (e *EmployeeReports) GetEmployeeReports(db *gorm.DB) error {
	err := db.Preload("PayPeriod").Find(&e.EmployeeReports).Error
	if err != nil {
		return err
	}

	return nil
}
