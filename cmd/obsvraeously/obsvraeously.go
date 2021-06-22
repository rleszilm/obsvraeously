package main

import (
	"context"

	"github.com/rleszilm/genms/log"
	"github.com/rleszilm/genms/service"
	rest_service "github.com/rleszilm/genms/service/rest"
	"github.com/rleszilm/genms/service/rest/healthcheck"
	"github.com/rleszilm/obsvraeously/cmd/obsvraeously/config"
	"github.com/rleszilm/obsvraeously/internal/avrae"
	"github.com/rleszilm/obsvraeously/internal/handler"
)

const (
	cfgPrefix = "obsvrae"
)

// Variables used for command line parameters
var (
	logs = log.NewChannel("main")
)

func main() {
	logs.Info("starting obsvraeously")
	logs.Info("parsing config")
	cfg, err := config.NewFromEnv(cfgPrefix)
	if err != nil {
		logs.Fatal(err)
	}

	logs.Info("setting log level:", cfg.LogLevel)
	lvl, ok := log.LevelFromString(cfg.LogLevel)
	if !ok {
		logs.Fatal(cfg.LogLevel)
	}
	log.SetLevel(lvl)

	// create service manager
	logs.Info("creating service manager")
	manager := service.NewManager()

	// avrae data
	logs.Info("creating avrae service")
	avdb := avrae.NewService(cfg.Discord)
	manager.Register(avdb)

	// server
	logs.Info("creating rest server")
	rest, err := rest_service.NewServer("obsvraeously", &cfg.Rest)
	if err != nil {
		logs.Fatal(err)
	}
	manager.Register(rest)

	// health service
	logs.Info("creating health service")
	health := healthcheck.NewService(&cfg.Health, rest)
	manager.Register(health)

	// handler
	logs.Info("creating handler")
	handler := handler.NewService(avdb, rest)
	manager.Register(handler)

	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		logs.Fatal(err)
	}

	manager.Wait()
}
