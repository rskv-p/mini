package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"

	"github.com/rskv-p/mini/servs/s_nats/nats_cfg"
	"github.com/rskv-p/mini/servs/s_nats/nats_serv"
)

func main() {
	// Load config
	cfgData, err := os.ReadFile("_data/cfg/nats.config.json")
	if err != nil {
		panic("failed to read nats.config.json: " + err.Error())
	}

	var cfg nats_cfg.NatsConfig
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		panic("invalid config.json: " + err.Error())
	}

	// Init and start service
	svc := nats_serv.New(cfg)
	if err := svc.Init(); err != nil {
		panic("init failed: " + err.Error())
	}
	if err := svc.Start(); err != nil {
		panic("start failed: " + err.Error())
	}

	// Wait for termination
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-ctx.Done()
	_ = svc.Stop()
}
