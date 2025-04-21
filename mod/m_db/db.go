package m_db

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// DB Module
//---------------------

// DBModule represents the database module.
type DBModule struct {
	Client typ.IDBClient // Using the IDBClient interface for database operations
	Tables []interface{} // List of models to be handled by the module
}

// Ensure DBModule implements the IModule interface.
var _ typ.IModule = (*DBModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *DBModule) Name() string {
	return "m_db"
}

// Init initializes the database connection and creates the tables.
func (m *DBModule) Init() error {
	client := m.Client
	if client == nil {
		x_log.RootLogger().Errorf("Database client is not initialized")
		return fmt.Errorf("Database client is not initialized")
	}

	// Get the DB connection
	_, err := client.GetDB()
	if err != nil {
		x_log.RootLogger().Errorf("Failed to get DB connection: %v", err)
		return err
	}

	// Auto-migrate tables for all models handled by the module
	for _, model := range m.Tables {
		if err := client.Migrate(model); err != nil {
			x_log.RootLogger().Errorf("Failed to migrate table for model %v: %v", model, err)
			return err
		}
	}

	x_log.RootLogger().Info("Database initialized successfully with tables:", m.Tables)
	return nil
}

// Stop stops the module and closes the database connection (if necessary).
func (m *DBModule) Stop() error {
	x_log.RootLogger().Info("Stopping DB module")
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions the module can perform.
func (m *DBModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
		{
			Name: "db.create_record", // Create a new record in the database
			Func: func(a typ.IAction) any {
				action, ok := a.(*act.Action)
				if !ok {
					return fmt.Errorf("invalid action type")
				}

				modelName := action.InputString(0)
				data := action.Inputs[1] // Data to be used for creating the record

				// Find the model by its type and create a record in the database
				var model interface{}
				for _, table := range m.Tables {
					if fmt.Sprintf("%T", table) == modelName {
						model = table
					}
				}

				if model == nil {
					return fmt.Sprintf("Model %s not found", modelName)
				}

				// Convert data to the model type and create it
				if err := m.Client.Create(&data); err != nil {
					x_log.RootLogger().Errorf("Failed to create record: %v", err)
					return fmt.Errorf("failed to create record: %v", err)
				}

				return data
			},
		},
		{
			Name: "db.get_record", // Get a record from the database
			Func: func(a typ.IAction) any {
				action, ok := a.(*act.Action)
				if !ok {
					return fmt.Errorf("invalid action type")
				}

				modelName := action.InputString(0)
				id := action.InputInt(1)

				var model interface{}
				for _, table := range m.Tables {
					if fmt.Sprintf("%T", table) == modelName {
						model = table
					}
				}

				if model == nil {
					return fmt.Sprintf("Model %s not found", modelName)
				}

				if err := m.Client.Find(&model, id); err != nil {
					return fmt.Sprintf("Failed to find record: %v", err)
				}

				return model
			},
		},
	}
}

//---------------------
// Model Management
//---------------------

// AddModel adds a new model to the DB module.
func (m *DBModule) AddModel(model interface{}) {
	// Ensure the model is not nil
	if model == nil {
		x_log.RootLogger().Error("Cannot add nil model")
		return
	}
	m.Tables = append(m.Tables, model)
}
