package main

import (
	log "k8s.io/klog/v2"
	http "github.com/capital-prawn/capper/http"
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	// core "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
var (
port string
)

func main() {
	// flag.StringVar(&tlscert, "tlscert", "/etc/certs/tls.crt", "Path to the TLS certificate")
	// flag.StringVar(&tlskey, "tlskey", "/etc/certs/tls.key", "Path to the TLS key")
	flag.StringVar(&port, "port", "8443", "The port to listen")
	flag.Parse()

	server := http.NewServer(port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Errorf("Failed to listen and serve: %v", err)
		}
	}()

	log.Infof("Server running in port: %s", port)

	// listen shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Infof("Shutdown gracefully...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Error(err)
	}
}