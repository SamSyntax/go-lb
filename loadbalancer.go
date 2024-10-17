package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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
	mu      sync.Mutex
	alive   bool
}

func (s *LbServer) Address() string {
	return s.addr
}

func (s *LbServer) IsAlive() bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get(s.addr)
	if err != nil {
		log.Warnf("Server %s - addr: %s is currently offline\n", s.name, s.addr)
		s.alive = false
		return false
	}
	if res.StatusCode != http.StatusOK {
		log.Warnf("Server %s - addr: %s is currently offline\n", s.name, s.addr)
		s.alive = false
		return false
	}
	s.alive = true
	log.Infof("Server %s - addr: %s is online\n", s.name, s.addr)
	return true
}

func (s *LbServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func NewLbServer(addr string, weight int) *LbServer {
	serverUrl, err := url.Parse(addr)
	if err != nil {
		log.Panicf("Failed to parse url: %v\n", err)
		os.Exit(1)
	}
	return &LbServer{
		addr:   addr,
		proxy:  httputil.NewSingleHostReverseProxy(serverUrl),
		weight: weight,
		alive:  true,
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []*LbServer
	weighted        bool
	mu              sync.Mutex
}

func NewLoadBalancer(port string, servers []*LbServer, weighted bool) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
		weighted:        weighted,
	}
}

func (lb *LoadBalancer) GetNextAvailableServer() LbServer {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if lb.weighted {
		return *lb.getWeightedServer()
	}
	return *lb.getRoundRobinServer()
}
func (lb *LoadBalancer) getWeightedServer() *LbServer {

	totalServers := len(lb.servers)
	for i := 0; i < totalServers; i++ {
		server := lb.servers[lb.roundRobinCount%totalServers]
		if server.current < server.weight && server.alive {
			server.mu.Lock()
			server.current++
			server.mu.Unlock()
			return server
		}
		server.current = 0
		lb.roundRobinCount++
	}
	lb.roundRobinCount++
	return lb.servers[lb.roundRobinCount%totalServers]
}
func (lb *LoadBalancer) getRoundRobinServer() *LbServer {
	totalServers := len(lb.servers)
	for i := 0; i < totalServers; i++ {
		server := lb.servers[lb.roundRobinCount%len(lb.servers)]
		if server.alive {
			lb.roundRobinCount++
			return server
		}
		lb.roundRobinCount++
	}
	return lb.servers[lb.roundRobinCount%len(lb.servers)]
}

func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.GetNextAvailableServer()
	log.Infof("forwarding to %q\n", targetServer.Address())
	targetServer.Serve(w, r)
}

func (lb *LoadBalancer) HealthCheck(interval time.Duration) {
  for _, server := range lb.servers {
    go func(s *LbServer) {
      ticker := time.NewTicker(interval)
      defer ticker.Stop()
      for  {
        <-ticker.C
        s.IsAlive()
      }
    }(server)
  }
}
