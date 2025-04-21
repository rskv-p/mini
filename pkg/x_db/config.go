package x_db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//---------------------
// Database Config
//---------------------

// DatabaseConfig contains the configuration parameters for connecting to the database.
type DatabaseConfig struct {
	Dialect  string // Database dialect (e.g., sqlite, mysql, postgres)
	Host     string // Host address
	Port     int    // Port number
	User     string // Database username
	Password string // Database password
	DbName   string // Database name
}

//---------------------
// Database Initialization
//---------------------

// InitDB initializes the database connection using GORM.
func InitDB(config DatabaseConfig) (*gorm.DB, error) {
	// Set up the connection string (DSN) for the database
	dsn := "gorm.db" // Configure the connection string for different databases (e.g., MySQL, PostgreSQL, SQLite)

	// Open a connection to the database
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err // Return error if the connection fails
	}
	return db, nil // Return the database instance if successful
}
