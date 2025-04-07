package cmd_runn

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/servs/s_runn/runn_cfg"
	"github.com/rskv-p/mini/servs/s_runn/runn_client"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"

	"github.com/spf13/cobra"
)

var force bool

var Cmd = &cobra.Command{
	Use:   "runn",
	Short: "Run service launcher",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start services via launcher",
	RunE: func(cmd *cobra.Command, args []string) error {
		log, _ := x_log.NewLogger()

		data, err := os.ReadFile("_data/cfg/runn.config.json")
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}

		var cfg runn_cfg.RunnConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}

		ordered, err := runn_cfg.ResolveStartupOrder(cfg)
		if err != nil {
			return fmt.Errorf("resolve order: %w", err)
		}

		launcher := runn_serv.New(cfg)
		client := runn_client.NewLocalClient(launcher)

		state, _ := runn_serv.LoadState()
		active := map[string]runn_serv.StateEntry{}
		if state != nil {
			active = state.Processes
		}

		if force {
			log.Infow("--force: stopping all")
			_ = client.StopAllServices(context.Background())
		}

		for _, svc := range ordered {
			if !svc.AutoRestart {
				continue
			}
			if prev, ok := active[svc.Name]; ok && !force {
				log.Infow("already running, skip", "name", svc.Name, "pid", prev.Pid)
				continue
			}
			if err := client.StartService(context.Background(), svc.Name); err != nil {
				log.Errorw("failed to start", "name", svc.Name, "err", err)
			} else {
				log.Infow("started", "name", svc.Name)
			}
		}

		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all running services",
	RunE: func(cmd *cobra.Command, args []string) error {
		log, _ := x_log.NewLogger()
		data, err := os.ReadFile("_data/cfg/runn.config.json")
		if err != nil {
			return err
		}

		var cfg runn_cfg.RunnConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return err
		}

		launcher := runn_serv.New(cfg)
		client := runn_client.NewLocalClient(launcher)

		log.Infow("stopping all services")
		return client.StopAllServices(context.Background())
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List running services",
	RunE: func(cmd *cobra.Command, args []string) error {
		log, _ := x_log.NewLogger()
		data, err := os.ReadFile("_data/cfg/runn.config.json")
		if err != nil {
			return err
		}

		var cfg runn_cfg.RunnConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return err
		}

		launcher := runn_serv.New(cfg)
		client := runn_client.NewLocalClient(launcher)

		list, err := client.List(context.Background())
		if err != nil {
			return err
		}

		for _, svc := range list {
			status := "stopped"
			if svc.Running {
				status = fmt.Sprintf("running (pid=%d, uptime=%ds)", svc.Pid, svc.Uptime)
			}
			log.Infow("service", "name", svc.Name, "status", status)
		}
		return nil
	},
}

func init() {
	startCmd.Flags().BoolVar(&force, "force", false, "Stop all and restart")
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(listCmd)
}
