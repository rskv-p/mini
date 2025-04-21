package x_db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//---------------------
// Database Config
//---------------------

// DatabaseConfig содержит параметры конфигурации для подключения к базе данных.
type DatabaseConfig struct {
	Dialect  string // Драйвер базы данных (например, sqlite, mysql, postgres)
	Host     string // Адрес хоста
	Port     int    // Номер порта
	User     string // Имя пользователя базы данных
	Password string // Пароль базы данных
	DbName   string // Имя базы данных
}

//---------------------
// Database Initialization
//---------------------

// InitDB инициализирует соединение с базой данных с использованием GORM и конфигурации.
func InitDB(config DatabaseConfig) (*gorm.DB, error) {
	var dsn string
	// Настройка строки подключения (DSN) для разных баз данных
	switch config.Dialect {
	case "sqlite":
		// Для SQLite достаточно имени файла базы данных
		dsn = config.DbName
	case "mysql":
		// Строка подключения для MySQL
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.User, config.Password, config.Host, config.Port, config.DbName)
	case "postgres":
		// Строка подключения для PostgreSQL
		dsn = fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", config.Host, config.Port, config.User, config.DbName, config.Password)
	default:
		return nil, fmt.Errorf("unsupported database dialect: %s", config.Dialect)
	}

	// Открытие соединения с базой данных в зависимости от диалекта
	var db *gorm.DB
	var err error
	switch config.Dialect {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil // Возвращаем экземпляр базы данных, если соединение успешно
}
