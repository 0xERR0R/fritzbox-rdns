package main

import (
	"github.com/0xERR0R/fritzbox-rdns/cache"
	"github.com/0xERR0R/fritzbox-rdns/config"
	"github.com/0xERR0R/fritzbox-rdns/fritzbox"
	"github.com/0xERR0R/fritzbox-rdns/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initializeLogging()

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal().Err(err).Msg("configuration error")
	}

	service := fritzbox.NewService(cfg.Url, cfg.User, cfg.Password)

	c := cache.NewCache(service)

	srv, err := server.NewServer(c)

	if err != nil {
		log.Fatal().Err(err).Msg("can't create DNS server")
	}

	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	srv.Start()

	go func() {
		<-signals
		log.Info().Msg("Terminating...")
		srv.Stop()
		done <- true
	}()

	<-done
}

func initializeLogging() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "2006-01-02T15:04:05"
	})
	log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "2006-01-02T15:04:05"
	}))
}
