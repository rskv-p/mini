package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/servs/s_nats/nats_cfg"
	"github.com/rskv-p/mini/servs/s_nats/nats_serv"
)

func main() {

	cfg, err := nats_cfg.LoadConfig()
	if err != nil {
		x_log.Error().Err(err).Str("file", "nats.config.json").Msg("failed to read config")
		os.Exit(1)
	}

	x_log.InitWithConfig(&cfg.Logger, "nats")

	// Init and start service
	svc := nats_serv.New(*cfg)
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
