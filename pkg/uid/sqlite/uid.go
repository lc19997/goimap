package sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// UIDEntry is the GORM model that represents an ID.
type UIDEntry struct {
	gorm.Model
	ExternalID string `gorm:"uniqueIndex"`
}

type UIDMapper struct {
	db *gorm.DB
}

// New initializes a new UIDMapper with a Gorm SQLite backend.
func New(databaseName string) (*UIDMapper, error) {
	db, err := gorm.Open(sqlite.Open(databaseName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// AutoMigrate will create the UIDEntry table if it does not exist.
	err = db.AutoMigrate(&UIDEntry{})
	if err != nil {
		return nil, err
	}

	return &UIDMapper{db: db}, nil
}

// FindOrAdd implements the FindOrAdd method of the UIDMapper interface.
func (u *UIDMapper) FindOrAdd(externalID string) uint32 {
	var entry UIDEntry
	result := u.db.Where(UIDEntry{ExternalID: externalID}).FirstOrCreate(&entry)
	if result.Error != nil {
		// Handle error appropriately.
		panic(result.Error)
	}

	// Convert the ID to uint32 since GORM uses uint as the default primary key type.
	return uint32(entry.ID)
}

// Remove implements the Remove method of the UIDMapper interface.
func (u *UIDMapper) Remove(externalID string) {
	u.db.Where("external_id = ?", externalID).Unscoped().Delete(&UIDEntry{})
}

func (u *UIDMapper) Validity() uint32 {
	return 1
}
