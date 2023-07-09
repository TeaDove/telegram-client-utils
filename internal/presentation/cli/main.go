package cli

import (
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/container"
	"github.com/teadove/goteleout/internal/utils"
	"github.com/urfave/cli"
)

var combatContainer container.Container

func action(cCtx *cli.Context) error {
	err := combatContainer.Presentation.Run()
	return err
}

func init() {
	combatContainer = container.MustNewCombatContainer()
}

func RunCli() {
	captureInterrupt()
	log.Info().
		Str("status", "starting.application").
		Msgf("All sessions will be stored in %s", combatContainer.TelegramSessionStorageFullPath)

	var flags []cli.Flag
	app := &cli.App{
		Name:      "telegram-client-utils",
		Usage:     "",
		UsageText: "",
		Flags:     flags,
		Action:    action,
	}

	err := app.Run(os.Args)
	utils.Check(err)
}

func captureInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			println()
			log.Info().Str("signal", sig.String()).Msg("captured exit signal, exiting...")
			pprof.StopCPUProfile()
			os.Exit(0)
		}
	}()
}
