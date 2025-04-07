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
	force := len(os.Args) > 1 && strings.EqualFold(os.Args[1], "--force")

	// Load config
	data, err := os.ReadFile("_data/cfg/runn.config.json")
	if err != nil {
		x_log.Error().Err(err).Str("file", "runn.config.json").Msg("failed to read config")
		os.Exit(1)
	}

	var cfg runn_cfg.RunnConfig
	x_log.InitWithConfig(&cfg.Logger, "runn")
	if err := json.Unmarshal(data, &cfg); err != nil {
		x_log.Error().Err(err).Msg("failed to parse config")
		os.Exit(1)
	}

	x_log.Info().Int("services", len(cfg.Services)).Msg("config loaded")

	// Resolve startup order
	ordered, err := runn_cfg.ResolveStartupOrder(cfg)
	if err != nil {
		x_log.Error().Err(err).Msg("dependency resolution failed")
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
		x_log.Info().Msg("--force mode: stopping all previously running services")
		_ = client.StopAllServices(context.Background())
	}

	// Start services in resolved order
	for _, svc := range ordered {
		if !svc.AutoRestart {
			continue
		}

		if prev, ok := active[svc.Name]; ok && !force {
			x_log.Info().Str("name", svc.Name).Int("pid", prev.Pid).Msg("already running, skipping")
			continue
		}

		if err := client.StartService(context.Background(), svc.Name); err != nil {
			x_log.Error().Err(err).Str("name", svc.Name).Msg("failed to start service")
		} else {
			x_log.Info().Str("name", svc.Name).Msg("service started")
		}
	}

	// Wait for interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	x_log.Info().Msg("runn active — press Ctrl+C to stop")
	<-ctx.Done()

	x_log.Info().Msg("shutting down all services")
	_ = client.StopAllServices(context.Background())
	x_log.Info().Msg("done")
}
