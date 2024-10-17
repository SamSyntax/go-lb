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
	addr    string
	proxy   *httputil.ReverseProxy
	name    string
	weight  int
	current int
}

func (s LbServer) Address() string {
	return s.addr
}

func (s LbServer) IsAlive() bool {
	res, err := http.Get(s.addr)
	if err != nil {
		fmt.Printf("Server %s - addr: %s is currently offline\n", s.name, s.addr)
		return false
	}
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Server %s - addr: %s is currently offline\n", s.name, s.addr)
		return false
	}
	fmt.Printf("Server %s - addr: %s is online\n", s.name, s.addr)
	return true
}

func (s LbServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func NewLbServer(addr string, weight int) *LbServer {
	serverUrl, err := url.Parse(addr)
	if err != nil {
		fmt.Printf("Failed to parse url: %v\n", err)
		os.Exit(1)
	}
	return &LbServer{
		addr:   addr,
		proxy:  httputil.NewSingleHostReverseProxy(serverUrl),
		weight: weight,
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []LbServer
	weighted        bool
}

func NewLoadBalancer(port string, servers []LbServer, weighted bool) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
		weighted:        weighted,
	}
}

func (lb *LoadBalancer) GetNextAvailableServer() LbServer {
	if lb.weighted {
		return lb.getWeightedServer()
	}
	return lb.getRoundRobinServer()
}
func (lb *LoadBalancer) getWeightedServer() LbServer {

	totalServers := len(lb.servers)
	for i := 0; i < totalServers; i++ {
		server := &lb.servers[lb.roundRobinCount%totalServers]
		if server.current < server.weight && server.IsAlive() {
			server.current++
			return *server
		}
		server.current = 0
		lb.roundRobinCount++
	}
	lb.roundRobinCount++
	return lb.servers[lb.roundRobinCount%totalServers]
}
func (lb *LoadBalancer) getRoundRobinServer() LbServer {

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
