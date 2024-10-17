package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

func main() {
	// creating flags to pass args
	nFlag := flag.Int("amount", 5, "Enter amount of local servers to be spawned")
	method := flag.String("method", "rr", "Load balancing method: 'rr - Round Robin | wrr - Weighted Round Robin")
	env := flag.String("env", "local", "Specify wether local servers should be started or provide JSON file with addresses of external servers. ")
	path := flag.String("path", "./servers.yaml", "Specify a path to servers config file. Either yaml or json. ")
	Lbport := flag.String("port", "7000", "Specify port on which load balancer is launched.")
	LocalServerPort := flag.Int("srv-port", 8000, "Specify port on which local dev server is launched. (If there are more than 1 server to be launched port number will be incremented by 1 for every local server to be spawned ex. server 0 - :8000; sever 1 - :8001)")
	healthCheck := flag.Bool("healthCheck", false, "Run health check on external servers from the list")

	// parsing flags
	flag.Parse()

	var lb *LoadBalancer
	var servers []*LbServer
	if *healthCheck {
		*env = "external"
	}
	switch *env {
	case "external":
		servers = Loader(*path)
	case "local":
		servers = Spawner(*nFlag, *LocalServerPort)
	default:
		servers = Spawner(*nFlag, *LocalServerPort)
	}

	switch *method {
	case "wrr":
		lb = NewLoadBalancer(*Lbport, servers, true)
	case "rr":
		lb = NewLoadBalancer(*Lbport, servers, false)
	default:
		log.WithFields(log.Fields{
			"method": *method,
		}).Fatalf("Invalid method. Use 'rr' or 'wrr', got %s", *method)
		os.Exit(1)
	}

	if *healthCheck && *path != "" {
		lb.HealthCheck(1 * time.Second)
		os.Exit(1)
	}
	lb.HealthCheck(5 * time.Second)
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(w, r)
	}

	// Log aggregation
	file, err := os.OpenFile("application.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer file.Close()
	log.SetOutput(file)
	// Serving loadbalancer
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("serving requests at localhost:%s\n", lb.port)
	log.WithFields(log.Fields{
		"port": lb.port,
	}).Printf("serving requests at localhost:%s\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)

	select {}
}
