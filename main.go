// main.go
package main

import (
	"context"
	"flag"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// Define command-line flags
var (
	amount              = flag.Int("amount", 5, "Enter amount of local servers to be spawned")
	method              = flag.String("method", "rr", "Load balancing method: 'rr' - Round Robin | 'wrr' - Weighted Round Robin")
	env                 = flag.String("env", "local", "Specify whether local servers should be started or provide JSON file with addresses of external servers.")
	path                = flag.String("path", "./servers.yaml", "Specify a path to servers config file. Either yaml or json.")
	lbPort              = flag.Int("port", 7000, "Specify port on which load balancer is launched.")
	serverPort          = flag.Int("srv-port", 8000, "Specify port on which local dev server is launched. (If more than 1 server is spawned, the port number will be incremented by 1 for each server, e.g., server 0 - :8000; server 1 - :8001)")
	healthCheck         = flag.Bool("healthCheck", false, "Run health check on external servers from the list")
	healthCheckInterval = flag.Int("hcInterval", 20, "Specify interval between running health checks on servers in the pool")
	configPath          = flag.String("config", "./config.json", "Specify a path to balancer config file in json format")
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetLevel(log.TraceLevel)
}

func flagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	// Parse flags
	flag.Parse()

	// Load configuration from JSON file
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	// Override config with flags if flags were set
	if flagPassed("amount") {
		cfg.Amount = *amount
	}
	if flagPassed("method") {
		cfg.Method = *method
	}
	if flagPassed("env") {
		cfg.Environment = *env
	}
	if flagPassed("path") {
		log.Infof("Flag 'path' is set to %s, but handling is not implemented in this example.", *path)
	}
	if flagPassed("port") {
		cfg.Balanceer_port = *lbPort
	}
	if flagPassed("srv-port") {
		cfg.Servers_port = *serverPort
	}
	if flagPassed("hcInterval") {
		cfg.Health_check_interval = *healthCheckInterval
	}

	// Initialize OpenTelemetry exporter
	epInit()
	ctx := context.Background()
	exp, err := newOTLPExporter(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize exporter: %v", err)
	}

	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("go-lb")

	// Define LoadBalancer
	var lb *LoadBalancer
	var servers []*LbServer

	// Check for environment
	switch cfg.Environment {
	case "external":
		if len(cfg.Servers) > 0 {
			servers, err = LoadServers(cfg.Servers)
			if err != nil {
				log.Errorf("Error loading external servers: %v", err)
			}
		} else {
			servers, err = Loader(*path)
			if err != nil {
				log.Errorf("Error loading servers from path: %v", err)
			}
		}
	case "local":
		servers = Spawner(cfg.Amount, cfg.Servers_port)
	default:
		log.Fatalf("Unknown environment: %s", cfg.Environment)
	}

	// Check for balancing method
	switch cfg.Method {
	case "wrr":
		lb = NewLoadBalancer(cfg.Balanceer_port, servers, true)
	case "rr":
		lb = NewLoadBalancer(cfg.Balanceer_port, servers, false)
	default:
		log.WithFields(log.Fields{
			"method": cfg.Method,
		}).Fatalf("Invalid method. Use 'rr' or 'wrr', got %s", cfg.Method)
	}

	if *healthCheck && cfg.Environment == "external" {
		lb.HealthCheck(1 * time.Second)
		os.Exit(0)
	} else {
		lb.HealthCheck(time.Duration(cfg.Health_check_interval) * time.Second)
	}

	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "HTTP GET /")
		defer span.End()

		lb.ServeProxy(w, r, ctx)
	}

	// Log aggregation
	file, err := os.OpenFile("application.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer file.Close()
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	// Serving load balancer
	http.HandleFunc("/", handleRedirect)
	log.WithFields(log.Fields{
		"port":    lb.port,
		"address": "127.0.0.1",
	}).Print("Serving requests at\n")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(lb.port), nil))
}
