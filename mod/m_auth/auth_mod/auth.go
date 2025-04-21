package auth_mod

import (
	"fmt"

	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_db/db_type"
	"github.com/rskv-p/mini/mod/m_log/log_type"
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
	Module    typ.IModule         // The module implementing IModule interface
	Client    db_type.IDBClient   // Database client for operations
	logClient log_type.ILogClient // Client for logging
}

// Ensure MAuth implements the IModule interface.
var _ typ.IModule = (*MAuth)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *MAuth) GetName() string {
	return m.Module.GetName()
}

// Start starts the module (specific logic can be added).
func (s *MAuth) Start() error {
	// Implement start logic if needed
	return nil
}

// Init initializes the module (specific logic can be added).
func (s *MAuth) Init() error {
	// Implement initialization logic if needed
	return nil
}

// Stop stops the module (useful for resource cleanup).
func (m *MAuth) Stop() error {
	m.logClient.Info("Stopping MAuth module", map[string]interface{}{"module": m.GetName()})
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions the module can perform.
func (m *MAuth) GetActions() []act_type.ActionDef {
	return []act_type.ActionDef{
		{
			Name: "auth.register", // Register a new admin
			Func: func(a act_type.IAction) any {
				action, ok := a.(*act_core.Action)
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
					m.logClient.Error(fmt.Sprintf("Failed to hash password: %v", err), map[string]interface{}{"username": username, "error": err})
					return fmt.Errorf("Error hashing password: %v", err)
				}

				// Create a new admin
				admin := Admin{
					Username:     username,
					PasswordHash: string(hash),
				}

				// Attempt to create the admin in the database
				m.Client.Create(&admin)

				// Log success
				m.logClient.Info("Admin created successfully", map[string]interface{}{"username": username})
				return "Admin created successfully"
			},
		},
		{
			Name: "auth.login", // Admin login action
			Func: func(a act_type.IAction) any {
				action, ok := a.(*act_core.Action)
				if !ok {
					return fmt.Errorf("invalid action type")
				}

				username := action.InputString(0)
				password := action.InputString(1)

				var admin Admin
				// Find the admin by username
				if err := m.Client.First(&admin, "username = ?", username).Error; err != nil {
					m.logClient.Error("Invalid username or password", map[string]interface{}{"username": username})
					return "Invalid username or password"
				}

				// Check the password
				if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
					m.logClient.Error("Invalid username or password", map[string]interface{}{"username": username})
					return "Invalid username or password"
				}

				// Successful login
				m.logClient.Info("Login successful", map[string]interface{}{"username": username})
				return fmt.Sprintf("Login successful for %s", username)
			},
		},
	}
}

//---------------------
// Module Creation
//---------------------

// NewMAuthModule creates a new instance of MAuth and initializes it.
func NewMAuthModule(service typ.IService, client db_type.IDBClient, logClient log_type.ILogClient) *MAuth {
	// Create a new module using NewModule
	module := mod.NewModule("m_auth", service, nil, nil, nil)

	// Create and return the MAuth module
	authModule := &MAuth{
		Module:    module,
		Client:    client,
		logClient: logClient,
	}

	// Register actions for the module
	for _, action := range authModule.GetActions() {
		act_core.Register(action.Name, action.Func)
	}

	return authModule
}
