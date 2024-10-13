package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	nFlag := flag.Int("amount", 1234, "Enter amount of local servers to be spawned")
	method := flag.String("method", "rr", "Load balancing method: 'rr - Round Robin | wrr - Weighted Round Robin")
	flag.Parse()
	servers := Spawner(*nFlag)
	var lb *LoadBalancer
	switch *method {
	case "wrr":
		lb = NewLoadBalancer("7000", servers, true)
	case "rr":
		lb = NewLoadBalancer("7000", servers, false)
	default:
		fmt.Println("Invalid method. Use 'rr' or 'wrr'.")
		os.Exit(1)
	}
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(w, r)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at localhost:%s\n", lb.port)

	http.ListenAndServe(":"+lb.port, nil)
	select {}
}
