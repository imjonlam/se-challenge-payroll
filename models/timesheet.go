// Package models provides gorm models and related structs
package models

import (
	"time"

	"gorm.io/gorm"
)

// TimeReport struct (model)
type TimeSheet struct {
	ID         uint      `gorm:"primaryKey"`
	Date       time.Time `gorm:"primaryKey"`
	EmployeeID uint      `gorm:"primaryKey"`
	Hours      float32
	PayGroupID string `gorm:"type:varchar(191)"`
	PayGroup   PayGroup
}

// Insert a new entry into the time_sheets table
func (t *TimeSheet) CreatetTimeSheet(db *gorm.DB) error {
	err := db.FirstOrCreate(t, TimeSheet{ID: t.ID, Date: t.Date, EmployeeID: t.EmployeeID}).Error
	if err != nil {
		return err
	}

	return nil
}
