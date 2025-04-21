package x_db

import (
	"github.com/rskv-p/mini/pkg/x_log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//---------------------
// DB Client
//---------------------

var dbInstance *gorm.DB

// DBClient is the concrete implementation of the IDBClient interface.
type DBClient struct {
	Db *gorm.DB // Database connection instance
}

//---------------------
// Database Methods
//---------------------

// GetDB returns the database connection. It initializes the connection if it doesn't exist.
func (c *DBClient) GetDB() (*gorm.DB, error) {
	// If the database instance is not initialized, establish a connection
	if c.Db == nil {
		db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			x_log.RootLogger().Errorf("Failed to connect to the database: %v", err)
			return nil, err
		}
		c.Db = db
	}
	return c.Db, nil
}

// Migrate performs the database migration for the provided models.
func (c *DBClient) Migrate(models ...interface{}) error {
	// Get the database connection
	db, err := c.GetDB()
	if err != nil {
		return err
	}
	// Perform the migration
	return db.AutoMigrate(models...)
}

// Create inserts a record into the database.
func (c *DBClient) Create(model interface{}) error {
	// Get the database connection
	db, err := c.GetDB()
	if err != nil {
		return err
	}
	// Insert the model into the database
	return db.Create(model).Error
}

// Find finds a record in the database that matches the given conditions.
func (c *DBClient) Find(model interface{}, conditions ...interface{}) error {
	// Get the database connection
	db, err := c.GetDB()
	if err != nil {
		return err
	}
	// Find the record that matches the conditions
	return db.First(model, conditions...).Error
}

// First finds the first record that matches the given conditions.
func (c *DBClient) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return c.Db.First(dest, conds...)
}
