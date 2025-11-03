package main

import (
	"flag"
	"time"

	"github.com/shanth1/gotools/conf"
	"github.com/shanth1/gotools/consts"
	"github.com/shanth1/gotools/ctx"
	"github.com/shanth1/gotools/flags"
	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/internal/app"
	"github.com/shanth1/loggate/internal/config"
)

type Flags struct {
	Env        string `flag:"env" default:"local" usage:"Environment ('local' | 'development' | 'production')"`
	Level      string `flag:"level" default:"info" usage:"Level of logging ('trace' | 'debug' | 'info' | 'warn')"`
	ConfigPath string `flag:"config" usage:"Path to the YAML config file"`
}

func main() {
	ctx, shutdownCtx, cancel, shutdownCancel := ctx.WithGracefulShutdown(10 * time.Second)
	defer cancel()
	defer shutdownCancel()

	// --- Сonfig ---
	logger := log.New()

	flagCfg := &Flags{}
	if err := flags.RegisterFromStruct(flagCfg); err != nil {
		logger.Fatal().Err(err).Msg("register flags")
	}
	flag.Parse()

	cfg := &config.Config{}
	if err := conf.Load(flagCfg.ConfigPath, cfg); err != nil {
		logger.Fatal().Err(err).Msg("load config")
	}

	logger = logger.WithOptions(log.WithConfig(log.Config{
		Level:      flagCfg.Env,
		App:        cfg.App,
		Service:    cfg.Service,
		Console:    flagCfg.Env != consts.EnvProd,
		JSONOutput: flagCfg.Env == consts.EnvProd,
	}))

	ctx = log.NewContext(ctx, logger)
	app.Run(ctx, shutdownCtx, cfg)
}
