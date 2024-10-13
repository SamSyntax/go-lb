package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func Server(port, name string, weight int) LbServer {
	srv := NewLbServer("http://localhost"+port, weight)
	srv.name = name

	mux := http.NewServeMux()

	handler := func(w http.ResponseWriter, r *http.Request) {
		inf := fmt.Sprintf("Serving on port %s", port)
		fmt.Println(inf)
		io.WriteString(w, inf)
	}

	mux.HandleFunc("/", handler)
	fmt.Printf("Spawning server: %s at %s\n", srv.name, srv.addr)

	go func() {
		err := http.ListenAndServe(port, mux)
		if err != nil {
			fmt.Printf("Error starting server on port %s: %v\n", port, err)
		}
	}()

	return *srv

}
func Spawner(amt int) []LbServer {
	servers := make([]LbServer, 0, amt)
	weights := []int{5, 2, 3}
	for i := 0; i < amt; i++ {
		k := 0
		if i > len(weights) {
			k = 0
		}
		k++
		name := fmt.Sprintf("Server %v", i)
		port := ":" + strconv.Itoa(8000+i)
		srv := Server(port, name, weights[k])
		servers = append(servers, srv)
	}
	return servers
}
