package models

import "gorm.io/gorm"

// UIDEntry is the GORM model that represents an ID.
type UIDEntry struct {
	gorm.Model
	ExternalID string `gorm:"uniqueIndex"`
}
