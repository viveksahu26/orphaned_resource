package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/viveksahu26/orphaned_resource/pkg/config"
	"github.com/viveksahu26/orphaned_resource/pkg/orphanresources"
	"github.com/viveksahu26/orphaned_resource/pkg/probes"
)

func init() {
	// prometheus.MustRegister(kubeversion.KubeVersionEndOfSupport)
	// prometheus.MustRegister(kubeversion.KubeVersionEndOfLife)
	prometheus.MustRegister(orphanresources.OrphanedResources)
	config.LoadConfig()
}

func main() {
	// created a tickerForOrphanedResource which continues to tick after every 1 minute
	tickerForOrphanedResource := time.NewTicker(config.OrphanedResourceDuration)
	defer tickerForOrphanedResource.Stop()

	// channel to mark completion
	done := make(chan bool)

	// await termination signals from OS on a channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	http.Handle("/metrics", promhttp.Handler())

	// readiness check
	http.HandleFunc("/readyz", probes.ReadinessProbe)
	go http.ListenAndServe(":8080", nil)

	// start runner in a separate goroutine which
	// listens for either tick from the ticker or a a signal to stop.
	// go runner(ticker, done)
	go runnerForOrphanedRes(tickerForOrphanedResource, done)
}

func runnerForOrphanedRes(ticker *time.Ticker, done <-chan bool) {
	for {
		select {
		case <-done:
			return

		// the countorphanedresource.SendOrphanedResourceMetricsAlertInfo()
		// function will get triggered after every 1 minute
		case <-ticker.C:
			errSendCountOrphanedResourceAlert := orphanresources.SendOrphanedResourceMetricsAlertInfo()
			if errSendCountOrphanedResourceAlert != nil {
				log.Println("error while sending Orphaned Resource metric", errSendCountOrphanedResourceAlert)
			}
		}
	}
}
