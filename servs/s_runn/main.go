package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"

	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/servs/s_runn/runn_cfg"
	"github.com/rskv-p/mini/servs/s_runn/runn_client"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"
)

func main() {
	log, _ := x_log.NewLogger()

	force := len(os.Args) > 1 && strings.EqualFold(os.Args[1], "--force")

	// Load config
	data, err := os.ReadFile("_data/cfg/runn.config.json")
	if err != nil {
		log.Errorw("failed to read config", "file", "runn.config.json", "err", err)
		os.Exit(1)
	}

	var cfg runn_cfg.RunnConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Errorw("failed to parse config", "err", err)
		os.Exit(1)
	}

	log.Infow("config loaded", "services", len(cfg.Services))

	// Resolve startup order
	ordered, err := runn_cfg.ResolveStartupOrder(cfg)
	if err != nil {
		log.Errorw("dependency resolution failed", "err", err)
		os.Exit(1)
	}

	// Init launcher
	launcher := runn_serv.New(cfg)
	client := runn_client.NewLocalClient(launcher)

	// Load previous state
	state, _ := runn_serv.LoadState()
	active := map[string]runn_serv.StateEntry{}
	if state != nil {
		active = state.Processes
	}

	// Handle --force (stop all previously started services)
	if force {
		log.Infow("--force mode: stopping all previously running services")
		_ = client.StopAllServices(context.Background())
	}

	// Start services in resolved order
	for _, svc := range ordered {
		if !svc.AutoRestart {
			continue
		}

		if prev, ok := active[svc.Name]; ok && !force {
			log.Infow("already running, skipping", "name", svc.Name, "pid", prev.Pid)
			continue
		}

		if err := client.StartService(context.Background(), svc.Name); err != nil {
			log.Errorw("failed to start service", "name", svc.Name, "err", err)
		} else {
			log.Infow("service started", "name", svc.Name)
		}
	}

	// Wait for interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log.Infow("runn active — press Ctrl+C to stop")
	<-ctx.Done()

	log.Infow("shutting down all services")
	_ = client.StopAllServices(context.Background())
	log.Infow("done")
}
