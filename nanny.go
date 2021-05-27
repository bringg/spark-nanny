package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	maxRetries   int = 3
	initialDelay int = 60
)

type nanny struct {
	appName   string
	clientset *kubernetes.Clientset
	dryRun    bool
	logger    zerolog.Logger
	namespace string
	netClient *http.Client
}

func newNanny(appName, namespace string, timeout int, clientset *kubernetes.Clientset, dryRun bool) *nanny {
	// normalize the appName, change to lowercase, trim spaces and replace '_' with '-'
	appName = strings.TrimSpace(strings.ReplaceAll(strings.ToLower(appName), "_", "-"))

	n := &nanny{
		appName:   appName,
		namespace: namespace,
		// add some contextual fields to the log to better understand what's happening
		logger:    log.With().Str("spark-application", appName).Logger(),
		clientset: clientset,
		netClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		dryRun: dryRun,
	}

	n.logger.Debug().Msgf("creating nanny with the following config: name %s, namespace %s, timout %d", appName, namespace, timeout)

	return n
}

func (n *nanny) poke() {
	pod, err := n.clientset.CoreV1().Pods(n.namespace).Get(context.TODO(), fmt.Sprintf("%s-driver", n.appName), v1.GetOptions{})
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
		res, err := n.netClient.Get(endpoint)

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

func (n *nanny) kill() {
	n.logger.Info().Msg("going to delete driver pod")

	if !n.dryRun {
		if err := n.clientset.CoreV1().Pods(n.namespace).Delete(context.TODO(), fmt.Sprintf("%s-driver", n.appName), v1.DeleteOptions{}); err != nil {
			n.logger.Error().Err(err).Msg("failed to delete driver pod")
		}
	}
}
