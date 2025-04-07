package runn_cfg

import "fmt"

type RunnConfig struct {
	Services []ServiceConfig `json:"services"` // List of services to launch
}

type ServiceConfig struct {
	Name        string   `json:"name"`         // Logical service name
	Path        string   `json:"path"`         // Path to binary or main.go
	Args        []string `json:"args"`         // Optional arguments
	AutoRestart bool     `json:"auto_restart"` // Restart on crash
	DependsOn   []string `json:"depends_on"`   // List of service names this depends on
}

// resolveStartupOrder returns services in topological order based on dependencies
func ResolveStartupOrder(cfg RunnConfig) ([]ServiceConfig, error) {
	graph := map[string][]string{}
	services := map[string]ServiceConfig{}

	for _, svc := range cfg.Services {
		services[svc.Name] = svc
		graph[svc.Name] = svc.DependsOn
	}

	visited := map[string]bool{}
	temp := map[string]bool{}
	result := []ServiceConfig{}

	var visit func(string) error
	visit = func(n string) error {
		if temp[n] {
			return fmt.Errorf("circular dependency detected on %q", n)
		}
		if !visited[n] {
			temp[n] = true
			for _, dep := range graph[n] {
				if _, ok := services[dep]; !ok {
					return fmt.Errorf("unknown dependency %q for service %q", dep, n)
				}
				if err := visit(dep); err != nil {
					return err
				}
			}
			visited[n] = true
			temp[n] = false
			result = append(result, services[n])
		}
		return nil
	}

	for name := range services {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}
