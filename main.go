package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "dirty"
)

func main() {
	apps := flag.String("apps", "", "comma separated list of spark app names to watch")
	debug := flag.Bool("debug", false, "set log level to debug")
	interval := flag.Int("interval", 30, "interval between health checks in seconds")
	timeout := flag.Int("timeout", 10, "network timeout in seconds")
	dryRun := flag.Bool("dry-run", false, "preforms all the checks and logic, but won't actually delete the driver pod in case of failure")
	namespace := flag.String("namespace", "spark", "spark apps namespace")
	webListenAddress := flag.String("listen-address", ":9164", "Address to listen on for web interface and telemetry")

	flag.Parse()

	log.Info().Msgf("starting spark-nanny, version: %s, build date: %s, git commit: %s", version, buildDate, commit)
	buildInfo.Set(1)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if len(*apps) == 0 {
		log.Fatal().Msg("no spark apps to watch were defined")
	}

	log.Info().Msg("starting job scheduler")
	s := gocron.NewScheduler(time.UTC)

	log.Info().Msg("registering nannies")
	opts := nannyOpts{*dryRun, *interval, *namespace, *timeout}

	if err := registerNannies(s, *apps, opts); err != nil {
		log.Fatal().Err(err).Msg("failed to register nanny")
	}

	// start the scheduler asynchronously
	log.Info().Msgf("watching spark apps: %s", *apps)
	s.StartAsync()

	// setup a simple web server to expose health checks
	log.Info().Msgf("starting server on %s", *webListenAddress)
	ws := newServer(*webListenAddress)
	ws.startAsync()

	// channel to gracefully shutdown the job scheduler and server
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	<-sigint

	log.Info().Msg("stopping job scheduler")
	s.Stop()

	log.Info().Msg("stopping server")
	ws.stop()
}
