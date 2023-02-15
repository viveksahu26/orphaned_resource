package probes

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/viveksahu26/orphaned_resource/pkg/config"
)

// ReadinessProbe checks Prometheus connectivity
func ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	// get the Prometheus URL from config
	pingURL := strings.TrimSuffix(config.PrometheusURL, "api/v1/query")
	promURL, err := url.Parse(pingURL)
	if err != nil {
		log.Println("error while parsing Prometheus API URL", err)
	}

	// do the HTTP GET request to the API
	resp, err := http.Get(promURL.String())
	if err != nil {
		log.Println("error while pinging Prometheus", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cannot connect to Prometheus\n"))
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	log.Println("Readiness probe successful with", resp.Status)
}
