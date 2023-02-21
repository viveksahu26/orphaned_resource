package main

import (
	"fmt"
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
	log.Println("Inside main")
	defer log.Println("Exit main")
	// created a tickerForOrphanedResource which continues to tick after every 1 minute
	log.Println("1")
	tickerForOrphanedResource := time.NewTicker(config.OrphanedResourceDuration)
	defer tickerForOrphanedResource.Stop()

	// channel to mark completion
	done := make(chan bool)

	log.Println("2")
	// await termination signals from OS on a channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	log.Println("3")
	http.Handle("/metrics", promhttp.Handler())

	// readiness check
	http.HandleFunc("/readyz", probes.ReadinessProbe)
	go http.ListenAndServe(":8080", nil)

	// start runner in a separate goroutine which
	// listens for either tick from the ticker or a a signal to stop.
	// go runner(ticker, done)
	log.Println("4")
	fmt.Println("tickerForOrphanedResource: ", *tickerForOrphanedResource)
	fmt.Println("done: ", done)
	go runnerForOrphanedRes(tickerForOrphanedResource, done)

	// blocks the main goroutine infinitely until user terminates the process
	sig := <-shutdown
	done <- true
	log.Printf("received %s, terminating application", sig)
	close(shutdown)
	close(done)
}

func runnerForOrphanedRes(ticker *time.Ticker, done <-chan bool) {
	fmt.Println("Inside runnerForOrphanedRes ")
	defer fmt.Println("Exit runnerForOrphanedRes")
	for {
		select {
		case <-done:
			fmt.Println("RETURN")
			return

		// the countorphanedresource.SendOrphanedResourceMetricsAlertInfo()
		// function will get triggered after every 1 minute
		case <-ticker.C:
			fmt.Println("TICKER")
			errSendCountOrphanedResourceAlert := orphanresources.SendOrphanedResourceMetricsAlertInfo()
			if errSendCountOrphanedResourceAlert != nil {
				log.Println("error while sending Orphaned Resource metric", errSendCountOrphanedResourceAlert)
			}
		}
	}
}
