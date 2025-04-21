package db_type

import "gorm.io/gorm"

// IDBClient defines methods for interacting with a database.
type IDBClient interface {
	// GetDB returns the database connection
	GetDB() (*gorm.DB, error)

	// Migrate applies the database migrations for the provided models
	Migrate(models ...interface{}) error

	// First finds the first record that matches the conditions
	First(dest interface{}, conds ...interface{}) *gorm.DB

	// Create inserts a new record into the database
	Create(model interface{}) error

	// Find retrieves records that match the conditions
	Find(model interface{}, conditions ...interface{}) error
}
