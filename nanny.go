package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	maxRetries   int = 3
	initialDelay int = 60
)

// nanny watchs a spark application and kills the driver pod
// if the application is unresponsive
type nanny struct {
	app       string
	dryRun    bool
	logger    zerolog.Logger
	namespace string

	kc *kubeClient
	nc *http.Client
}

// nannyOpts controls the nannies behavior
// note that the options are shared across all nannies
type nannyOpts struct {
	dryRun    bool
	interval  int
	namespace string
	timeout   int
}

// registerNannies creates a new nanny for each app and registers it with
// the gocron scheduler
func registerNannies(s *gocron.Scheduler, apps string, opts nannyOpts) error {
	kubeClient := newKubeClient()
	netClient := &http.Client{
		Timeout: time.Duration(opts.timeout) * time.Second,
	}

	for _, app := range strings.Split(apps, ",") {
		// normalize app name
		app = strings.TrimSpace(strings.ReplaceAll(strings.ToLower(app), "_", "-"))

		n := &nanny{
			app:       app,
			namespace: opts.namespace,
			logger:    log.With().Str("spark-application", app).Logger(),
			dryRun:    opts.dryRun,
			kc:        kubeClient,
			nc:        netClient,
		}

		n.logger.Debug().Msgf("creating nanny with the following config: spark app %s/%s, timout %ds", opts.namespace, app, opts.timeout)

		if _, err := s.Every(opts.interval).Seconds().SingletonMode().Do(n.poke); err != nil {
			return err
		}
	}

	return nil
}

// poke checks the spark application responsiveness
func (n *nanny) poke() {
	pod, err := n.kc.getDriverPod(n.app, n.namespace)
	if err != nil {
		n.logger.Warn().Err(err).Msg("")
		return
	}

	if pod.DeletionTimestamp != nil {
		n.logger.Debug().Msg("pod already has deletion timestamp, nothing to do here")
		return
	}

	// usually we only have the driver container inside the pod,
	// but the crd does support "sidecar" containers so make sure they're
	// all in ready state
	for _, s := range pod.Status.ContainerStatuses {
		if !s.Ready {
			n.logger.Debug().Msgf("container %s isn't ready yet", s.Name)
			return
		}

		if s.State.Running != nil {
			age := time.Since(s.State.Running.StartedAt.Time)

			// we give some grace period before starting checks
			if age < time.Duration(initialDelay)*time.Second {
				n.logger.Debug().Msgf("container %s is still in startup grace period, age %s", s.Name, age.String())
				return
			}
		}
	}

	endpoint := fmt.Sprintf("http://%s:4040/api/v1/applications", pod.Status.PodIP)

	for i := 1; i <= maxRetries; i++ {
		n.logger.Debug().Msgf("pinging %s (retry %d/%d)", endpoint, i, maxRetries)
		res, err := n.nc.Get(endpoint)
		if err != nil {
			n.logger.Warn().Err(err).Msg("")
			n.logger.Debug().Msgf("got error with type %T", err)
			switch e := err.(type) {
			case *url.Error:
				// if it's a timout or connection refused error, we skip to the next
				// loop iteration
				if e.Timeout() {
					n.logger.Debug().Msg("got timeout")
					continue
				}
				if strings.Contains(e.Err.Error(), "connection refused") {
					n.logger.Debug().Msg("got connection refused")
					continue
				}
				// if it's some other type of error return so we don't kill
				// the pod on some network related issue
				return
			default:
				n.logger.Debug().Msg("will not retry")
				return
			}
		}

		defer res.Body.Close()

		// happy path
		if res.StatusCode == http.StatusOK {
			n.logger.Debug().Msg("got ok from the driver pod")
			return
		}

		n.logger.Warn().Msgf("got status code %d", res.StatusCode)
		// this might be temporary, so we wait a bit here and retry
		time.Sleep(3 * time.Second)
		continue
	}

	n.kill()
}

// kill deletes the driver pod
func (n *nanny) kill() {
	n.logger.Info().Msg("going to delete driver pod")

	if !n.dryRun {
		if err := n.kc.deleteDriverPod(n.app, n.namespace); err != nil {
			n.logger.Error().Err(err).Msg("failed to delete driver pod")
		}
	}
}
