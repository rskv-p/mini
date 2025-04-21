package db_core

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//---------------------
// DBClient
//---------------------

var dbInstance *gorm.DB

// DBClient is the concrete implementation of the IDBClient interface.
type DBClient struct {
	Db     *gorm.DB        // Database connection instance
	Config *DatabaseConfig // Configuration for the database connection
}

//---------------------
// Database Methods
//---------------------

// NewDBClient creates a new instance of DBClient.
func NewDBClient(config *DatabaseConfig) (*DBClient, error) {
	client := &DBClient{
		Config: config,
	}

	// Initialize the DB connection
	db, err := client.GetDB()
	if err != nil {
		return nil, err
	}

	client.Db = db
	return client, nil
}

// GetDB returns the database connection. It initializes the connection if it doesn't exist.
func (c *DBClient) GetDB() (*gorm.DB, error) {
	if c.Db == nil {
		var dsn string
		// Set up the connection string (DSN) based on the configured dialect
		switch c.Config.Dialect {
		case "sqlite":
			dsn = c.Config.DbName
			db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
			if err != nil {
				//	x_log.RootLogger().Errorf("Failed to connect to SQLite database: %v", err)
				return nil, err
			}
			return db, nil
		case "mysql":
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.Config.User, c.Config.Password, c.Config.Host, c.Config.Port, c.Config.DbName)
			db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				//	x_log.RootLogger().Errorf("Failed to connect to MySQL database: %v", err)
				return nil, err
			}
			return db, nil
		case "postgres":
			dsn = fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", c.Config.Host, c.Config.Port, c.Config.User, c.Config.DbName, c.Config.Password)
			db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err != nil {
				//	x_log.RootLogger().Errorf("Failed to connect to PostgreSQL database: %v", err)
				return nil, err
			}
			return db, nil
		default:
			return nil, fmt.Errorf("unsupported database dialect: %s", c.Config.Dialect)
		}
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
