package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type ServerInterface interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type LbServer struct {
	addr  string
	proxy *httputil.ReverseProxy
  name string
}

func (s LbServer) Address() string {
	return s.addr
}

func (s LbServer) IsAlive() bool {
	return true
}

func (s LbServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func NewLbServer(addr string) *LbServer {
	serverUrl, err := url.Parse(addr)
	if err != nil {
		fmt.Printf("Failed to parse url: %v\n", err)
		os.Exit(1)
	}
	return &LbServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []LbServer
}

func NewLoadBalancer(port string, servers []LbServer) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func (lb *LoadBalancer) GetNextAvailableServer() LbServer {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}

	lb.roundRobinCount++
	return server
}

func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.GetNextAvailableServer()
	fmt.Printf("forwarding to %q\n", targetServer.Address())
	targetServer.Serve(w, r)
}
