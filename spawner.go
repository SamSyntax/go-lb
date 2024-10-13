package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func Server(port, name string) LbServer {
	srv := NewLbServer("http://localhost" + port)
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
	for i := 0; i < amt; i++ {
		name := fmt.Sprintf("Server %v", i)
		port := ":" + strconv.Itoa(8000+i)
		srv := Server(port, name)
		servers = append(servers, srv)
	}
	return servers
}
