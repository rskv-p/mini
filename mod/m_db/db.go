package m_db

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_db"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// DBModule
//---------------------

// DBModule manages database interactions, including creating, retrieving, and migrating records.
type DBModule struct {
	Module       typ.IModule       // The module implementing IModule interface
	Client       typ.IDBClient     // Client for working with the database
	Tables       []interface{}     // List of models to handle
	ConfigClient typ.IConfigClient // Client for working with configuration
	logClient    typ.ILogClient    // Client for logging
}

// Ensure DBModule implements the IModule interface.
var _ typ.IModule = (*DBModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *DBModule) Name() string {
	return m.Module.Name()
}

// Start starts the module (specific startup logic can be added).
func (m *DBModule) Start() error {
	// Log the start action
	m.logClient.Info("Starting DB module", map[string]interface{}{"module": m.Name()})
	return nil
}

// Stop stops the module and closes the database connection.
func (m *DBModule) Stop() error {
	m.logClient.Info("Stopping DB module", map[string]interface{}{"module": m.Name()})
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
					m.logClient.Error("Invalid action type", map[string]interface{}{"action": action})
					return fmt.Errorf("invalid action type")
				}

				modelName := action.InputString(0)
				data := action.Inputs[1] // Data for creating the record

				// Find the model by type name and create a record in the database
				var model interface{}
				for _, table := range m.Tables {
					if fmt.Sprintf("%T", table) == modelName {
						model = table
					}
				}

				if model == nil {
					m.logClient.Error(fmt.Sprintf("Model %s not found", modelName), map[string]interface{}{"modelName": modelName})
					return fmt.Sprintf("Model %s not found", modelName)
				}

				// Convert data to the model type and create the record
				if err := m.Client.Create(&data); err != nil {
					m.logClient.Error(fmt.Sprintf("Failed to create record: %v", err), map[string]interface{}{"model": modelName, "data": data})
					return fmt.Errorf("failed to create record: %v", err)
				}

				m.logClient.Info(fmt.Sprintf("Created record for model: %s", modelName), map[string]interface{}{"modelName": modelName, "data": data})
				return data
			},
		},
		{
			Name: "db.get_record", // Retrieve a record from the database
			Func: func(a typ.IAction) any {
				action, ok := a.(*act.Action)
				if !ok {
					m.logClient.Error("Invalid action type", map[string]interface{}{"action": action})
					return fmt.Errorf("invalid action type")
				}

				modelName := action.InputString(0)
				id := action.InputInt(1)

				var model interface{}
				// Find the model by type name
				for _, table := range m.Tables {
					if fmt.Sprintf("%T", table) == modelName {
						model = table
					}
				}

				if model == nil {
					m.logClient.Error(fmt.Sprintf("Model %s not found", modelName), map[string]interface{}{"modelName": modelName})
					return fmt.Sprintf("Model %s not found", modelName)
				}

				if err := m.Client.Find(&model, id); err != nil {
					m.logClient.Error(fmt.Sprintf("Failed to find record: %v", err), map[string]interface{}{"modelName": modelName, "id": id})
					return fmt.Sprintf("Failed to find record: %v", err)
				}

				m.logClient.Info(fmt.Sprintf("Retrieved record for model: %s", modelName), map[string]interface{}{"modelName": modelName, "id": id})
				return model
			},
		},
	}
}

//---------------------
// Module Creation
//---------------------

// Init initializes the database connection and performs table migrations.
func (m *DBModule) Init() error {
	// Load database configuration via IConfigClient
	config := m.ConfigClient.GetConfig("db_config").(map[string]interface{})
	if config == nil {
		m.logClient.Error("Database configuration is not found in x_cfg", map[string]interface{}{"config": config})
		return fmt.Errorf("database configuration is not found")
	}

	// Create database configuration
	dbConfig := &x_db.DatabaseConfig{
		Dialect:  config["Dialect"].(string),
		Host:     config["Host"].(string),
		Port:     int(config["Port"].(float64)), // Convert from float64 to int
		User:     config["User"].(string),
		Password: config["Password"].(string),
		DbName:   config["DbName"].(string),
	}

	// Create a database client using the configuration
	client, err := x_db.NewDBClient(dbConfig)
	if err != nil {
		m.logClient.Error(fmt.Sprintf("Failed to create DB client: %v", err), map[string]interface{}{"dbConfig": dbConfig})
		return err
	}

	m.Client = client

	// Get the database connection
	_, err = m.Client.GetDB()
	if err != nil {
		m.logClient.Error(fmt.Sprintf("Failed to get DB connection: %v", err), map[string]interface{}{"dbConfig": dbConfig})
		return err
	}

	// Automatically migrate tables for all models
	for _, model := range m.Tables {
		if err := m.Client.Migrate(model); err != nil {
			m.logClient.Error(fmt.Sprintf("Failed to migrate table for model %v: %v", model, err), map[string]interface{}{"model": model})
			return err
		}
	}

	m.logClient.Info("Database initialized successfully with tables", map[string]interface{}{"tables": m.Tables})
	return nil
}

//---------------------
// Model Management
//---------------------

// AddModel adds a new model to the database module.
func (m *DBModule) AddModel(model interface{}) {
	// Ensure the model is not nil
	if model == nil {
		m.logClient.Error("Cannot add nil model", map[string]interface{}{"model": model})
		return
	}
	m.Tables = append(m.Tables, model)
	m.logClient.Info("Added model", map[string]interface{}{"model": model})
}

//---------------------
// Module Creation
//---------------------

// NewDBModule creates a new instance of DBModule and initializes it.
func NewDBModule(service typ.IService, configClient typ.IConfigClient, logClient typ.ILogClient) *DBModule {
	// Create a new module using NewModule, passing name, service, actions, and nil for OnInit and OnStop
	module := mod.NewModule("db", service, nil, nil, nil)

	// Create the DBModule with the created module and the clients
	dbModule := &DBModule{
		Module:       module,
		ConfigClient: configClient,
		logClient:    logClient,
	}

	// Register actions for the DB module
	for _, action := range dbModule.Actions() {
		act.Register(action.Name, action.Func)
	}

	// Return the created DBModule
	return dbModule
}
