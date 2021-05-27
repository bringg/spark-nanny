package main

import (
	"flag"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	apps := flag.String("apps", "", "comma separated list of spark app names to watch")
	debug := flag.Bool("debug", false, "set log level to debug")
	interval := flag.Int("interval", 30, "interval between health checks in seconds")
	timeout := flag.Int("timeout", 10, "network timeout in seconds")
	dryRun := flag.Bool("dry-run", false, "prefroms all the checks and logic, but won't actually delete the driver pod in case of failure")
	namespace := flag.String("namespace", "spark", "spark apps namespace")

	flag.Parse()

	log.Info().Msgf("starting spark-nanny, version: %s, build date: %s", version, buildDate)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if len(*apps) == 0 {
		log.Fatal().Msg("no spark apps to watch were defined")
	}

	log.Info().Msg("loading kubeconfig")

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Info().Msg("creating kubernetes clientset")

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Info().Msg("starting job scheduler")

	s := gocron.NewScheduler(time.UTC)

	log.Info().Msg("registering nannies")
	for _, a := range strings.Split(*apps, ",") {
		nanny := newNanny(a, *namespace, *timeout, clientset, *dryRun)

		_, err := s.Every(*interval).Seconds().SingletonMode().Do(nanny.poke)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to register job")
		}
	}

	s.StartBlocking()
}
