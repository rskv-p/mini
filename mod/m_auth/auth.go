package m_auth

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/typ"

	"golang.org/x/crypto/bcrypt"
)

//---------------------
// Admin Model
//---------------------

// Admin represents the admin model in the database.
type Admin struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;size:255"`
	PasswordHash string `gorm:"size:255"`
}

//---------------------
// MAuth Module
//---------------------

// MAuth represents the authentication module.
type MAuth struct {
	Client typ.IDBClient // Using IDBClient interface for database access
}

// Ensure MAuth implements the IModule interface.
var _ typ.IModule = (*MAuth)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the module name.
func (m *MAuth) Name() string {
	return "m_auth"
}

// Init initializes the module, setting up the database connection and creating the admin table.
func (m *MAuth) Init() error {
	client := m.Client
	if client == nil {
		x_log.RootLogger().Errorf("Database client is not initialized")
		return fmt.Errorf("Database client is not initialized")
	}

	// Automatically migrate the admin table
	if err := client.Migrate(&Admin{}); err != nil {
		x_log.RootLogger().Errorf("Failed to migrate table for admin model: %v", err)
		return err
	}

	x_log.RootLogger().Info("MAuth module initialized successfully")
	return nil
}

// Stop stops the module (useful for cleaning up resources if needed).
func (m *MAuth) Stop() error {
	x_log.RootLogger().Info("Stopping MAuth module")
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns the list of actions the module can perform.
func (m *MAuth) Actions() []typ.ActionDef {
	return []typ.ActionDef{
		{
			Name: "auth.register", // Register a new admin
			Func: func(a typ.IAction) any {
				action, ok := a.(*act.Action)
				if !ok {
					return fmt.Errorf("invalid action type")
				}

				username := action.InputString(0)
				password := action.InputString(1)

				// Check if the username already exists
				var existingAdmin Admin
				if err := m.Client.First(&existingAdmin, "username = ?", username).Error; err == nil {
					return fmt.Sprintf("Username '%s' is already taken", username)
				}

				// Hash the password
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					x_log.RootLogger().Errorf("Failed to hash password: %v", err)
					return fmt.Errorf("Error hashing password: %v", err)
				}

				// Create a new admin
				admin := Admin{
					Username:     username,
					PasswordHash: string(hash),
				}

				// Attempt to create the admin in the database
				m.Client.Create(&admin)

				return "Admin created successfully"
			},
		},
		{
			Name: "auth.login", // Admin login action
			Func: func(a typ.IAction) any {
				action, ok := a.(*act.Action)
				if !ok {
					return fmt.Errorf("invalid action type")
				}

				username := action.InputString(0)
				password := action.InputString(1)

				var admin Admin
				// Find the admin by username
				if err := m.Client.First(&admin, "username = ?", username).Error; err != nil {
					return "Invalid username or password"
				}

				// Verify the password
				if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
					return "Invalid username or password"
				}

				// Successful login
				return fmt.Sprintf("Login successful for %s", username)
			},
		},
	}
}
