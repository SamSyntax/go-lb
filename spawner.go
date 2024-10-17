package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func Server(port, name string, weight int) *LbServer {
	srv := NewLbServer("http://localhost"+port, weight)
	srv.name = name

	mux := http.NewServeMux()

	handler := func(w http.ResponseWriter, r *http.Request) {
		inf := fmt.Sprintf("Serving on port %s", port)
		log.Info(inf)
		io.WriteString(w, inf)
	}

	mux.HandleFunc("/", handler)
	log.Infof("Spawning server: %s at %s\n", srv.name, srv.addr)

	go func() {
		err := http.ListenAndServe(port, mux)
		if err != nil {
			fmt.Printf("Error starting server on port %s: %v\n", port, err)
		}
	}()

	return srv

}
func Spawner(amt, port int) []*LbServer {
	servers := make([]*LbServer, 0, amt)
	weights := []int{5, 2, 3}
	for i := 0; i < amt; i++ {
		k := 0
		if i > len(weights) {
			k = 0
		}
		k++
		name := fmt.Sprintf("Server %v", i+1)
		port := ":" + strconv.Itoa(port+i)
		srv := Server(port, name, weights[k])
		if srv.IsAlive() {
			servers = append(servers, srv)
		} else {
			continue
		}

	}
	return servers
}
