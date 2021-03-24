// Package models provides gorm models and related structs
package models

import "gorm.io/gorm"

// PayGroup struct (model)
type PayGroup struct {
	ID   string `gorm:"type:varchar(191)"`
	Rate float32
}

// Insert a new record into pay_groups table
func (p *PayGroup) CreatePayGroup(db *gorm.DB) error {
	err := db.FirstOrCreate(&p).Error
	if err != nil {
		return err
	}

	return nil
}
