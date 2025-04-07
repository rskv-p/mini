package runn_serv

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"default:user"` // User roles: admin, user
	CreatedAt    time.Time
}

// CreateUser creates a new user with a hashed password
func CreateUser(db *gorm.DB, username, password, role string) error {
	// Validate input
	if username == "" || password == "" {
		return errors.New("username and password required")
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create the user in the database
	user := User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}
	return db.Create(&user).Error
}

// FindUserByUsername retrieves a user by their username
func FindUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CheckPassword compares the provided password with the stored hash
func (u *User) CheckPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pw)) == nil
}
