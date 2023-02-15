package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/viveksahu26/orphaned_resource/pkg/client"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// default values for config
var (
	defaultOrphanedResourceDuration = time.Minute
	defCertPath                     = "/opt/obmondo-k8s-agent/tls.crt"
	defKeyPath                      = "/opt/obmondo-k8s-agent/tls.key"
)

// External config values
var (
	OrphanedResourceDuration time.Duration
	PrometheusURL            string
	ObmondoURL               string
	ObmondoHTTPClient        *http.Client
	CertName                 string
	Clientset                *kubernetes.Clientset
	DynamicClient            *dynamic.DynamicClient
	DiscoveryClient          *discovery.DiscoveryClient
)

// LoadConfig populates the config vars from env
func LoadConfig() {
	OrphanedResourceDuration = getOrphanedResourceDuration()
	PrometheusURL = getPrometheusURL()
	ObmondoURL = getObmondoURL()
	cert := getCertificate()
	ObmondoHTTPClient = getCustomHTTPSClient(cert)
	CertName = getCommonNameForCert(cert)
	Clientset, DynamicClient = client.InitClient().Client()
	DiscoveryClient = Clientset.DiscoveryClient
}

// getDuration() loads the duration from env
func getOrphanedResourceDuration() time.Duration {
	rawDur := os.Getenv("ORPHANED_RES_DURATION")
	d, err := time.ParseDuration(rawDur)
	if err != nil {
		log.Println("unable to parse duration ", rawDur, err)
		log.Println("switching to default duration", defaultOrphanedResourceDuration)
		return defaultOrphanedResourceDuration
	}
	return d
}

// getPrometheusURL() loads the URL from env
func getPrometheusURL() string {
	promURL := os.Getenv("PROMETHEUS_URL")
	if len(promURL) == 0 {
		log.Fatal("unable to get Prometheus URL from env")
	}
	return fmt.Sprintf("%s/api/v1/query", promURL)
}

// getObmondoURL() loads the URL from env
func getObmondoURL() string {
	apiURL := os.Getenv("API_URL")
	if len(apiURL) == 0 {
		log.Fatal("unable to get Obmondo API URL from env")
	}
	return apiURL
}

// getCertificate() loads the certificates from env
func getCertificate() tls.Certificate {
	certPath := os.Getenv("AGENT_CERT_PATH")
	if len(certPath) == 0 {
		log.Println("unable to get cert path from env", certPath)
		log.Println("switching to default cert path", defCertPath)
		certPath = defCertPath
	}

	keyPath := os.Getenv("AGENT_KEY_PATH")
	if len(keyPath) == 0 {
		log.Println("unable to get key path from env", keyPath)
		log.Println("switching to default key path", defKeyPath)
		keyPath = defKeyPath
	}

	// Load client certificates
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatal("unable to read certificates")
	}

	return cert
}

// getCustomHTTPSClient creates an HTTPS client with the given certificates
func getCustomHTTPSClient(cert tls.Certificate) *http.Client {
	// Setup HTTPS client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	return &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
}

func getCommonNameForCert(cert tls.Certificate) string {
	if cert.Certificate == nil || len(cert.Certificate) == 0 {
		log.Fatal("Expected at least one certificate but found none when trying to get common name")
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		log.Fatalf("Could not parse certificate when trying to get common name: %v", err)
	}

	subject := fmt.Sprintf("%v", x509Cert.Subject)
	if len(subject) == 0 {
		log.Fatal("unable to read certificate's common name as the certificate's subject was empty")
	}

	certSubjectParts := strings.Split(subject, "=")
	if len(certSubjectParts) < 2 {
		log.Fatal("unable to read certificate's common name as less than two parts were found in the certificate subject")
	}

	return certSubjectParts[1]
}
