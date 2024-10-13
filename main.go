package main

import (
	"fmt"
	"net/http"
)

func main() {
	servers := Spawner(5)
	lb := NewLoadBalancer("7000", servers)
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(w, r)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at localhost:%s\n", lb.port)

	http.ListenAndServe(":"+lb.port, nil)
	select {}
}
