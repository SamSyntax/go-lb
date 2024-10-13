package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
)

type ExternalServer struct {
	Addr   string `json:"addr"`
	Weight int    `json:"weight"`
}

func ReadJson() []LbServer {
	file, err := os.Open("servers.json")
	if err != nil {
		log.Fatalf("Failed to load servers from JSON file %v", err)
	}
	defer file.Close()

	byteVal, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var servers []ExternalServer
	err = json.Unmarshal(byteVal, &servers)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	res := make([]LbServer, 0, len(servers))
	for k, s := range servers {
		lbServer := NewLbServer(s.Addr, s.Weight)
		lbServer.name = strconv.Itoa(k)
		res = append(res, *lbServer)
	}

	return res
}
