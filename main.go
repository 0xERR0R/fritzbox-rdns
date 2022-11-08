package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/0xERR0R/fritzbox-rdns/config"
	"github.com/0xERR0R/fritzbox-rdns/fritzbox"
	"github.com/0xERR0R/fritzbox-rdns/lookup"
	"github.com/0xERR0R/fritzbox-rdns/server"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal().Err(err).Msg("configuration error")
	}

	initializeLogging(cfg.LogLevel)

	service := fritzbox.NewService(cfg.Url, cfg.User, cfg.Password)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	c := lookup.NewNamesLookupService(service, rdb)

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

func initializeLogging(logLevel string) {
	level, err := zerolog.ParseLevel(logLevel)

	if err != nil {
		log.Fatal().Err(err).Msg("unknown log level")
	}

	zerolog.SetGlobalLevel(level)
	zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "2006-01-02T15:04:05"
	})
	log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "2006-01-02T15:04:05"
	}))
}
