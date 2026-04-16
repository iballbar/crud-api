package db

import (
	postgresadapter "crud-api/internal/adapters/postgres"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		postgresadapter.Models()...,
	)
}
