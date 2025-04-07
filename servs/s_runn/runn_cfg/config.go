package runn_cfg

import "time"

// PreconfiguredProcess defines a preconfigured process
type PreconfiguredProcess struct {
	Cmd           string `json:"cmd"`             // Path to the executable
	Dir           string `json:"dir"`             // Directory to run the process from
	Name          string `json:"name"`            // Process name
	StartOnLaunch bool   `json:"start_on_launch"` // Should the process start on launch
}

// Config holds all configurable fields for the s_proc service
type Config struct {
	DBPath                 string                 `json:"db_path"`                 // Path to the database file
	LogLevel               string                 `json:"log_level"`               // Logging level (e.g., info, debug)
	LogToFile              bool                   `json:"log_to_file"`             // Whether to log to a file
	HTTPAddress            string                 `json:"http_address"`            // HTTP server address
	RestartMax             int                    `json:"restart_max"`             // Maximum number of restarts
	Timeout                time.Duration          `json:"timeout_sec"`             // Timeout in seconds
	JwtSecret              string                 `json:"jwt_secret"`              // Secret key for JWT
	AdminDefaultPassword   string                 `json:"admin_password"`          // Default admin password
	AuthEnabled            bool                   `json:"auth_enabled"`            // Whether authentication is enabled
	PreconfiguredProcesses []PreconfiguredProcess `json:"preconfigured_processes"` // Preconfigured processes to be loaded
}

// Default configuration values
var defaultConfig = Config{
	DBPath:               "proc.db",
	LogLevel:             "info",
	LogToFile:            false,
	HTTPAddress:          ":8080",
	RestartMax:           3,
	Timeout:              10 * time.Second,
	JwtSecret:            "supersecret", // ⚠️ Change in production
	AdminDefaultPassword: "admin",       // ⚠️ Change in production
	AuthEnabled:          true,
	PreconfiguredProcesses: []PreconfiguredProcess{
		{
			Cmd:           "./microservice1",
			Dir:           "./microservices/",
			Name:          "microservice1",
			StartOnLaunch: true,
		},
		{
			Cmd:           "./microservice2",
			Dir:           "./microservices/",
			Name:          "microservice2",
			StartOnLaunch: false,
		},
	},
}

// Global config variable
var config Config

// C returns the current configuration
func C() Config {
	return config
}
