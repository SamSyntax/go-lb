package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ExternalServerJson struct {
	Addr   string `json:"address"`
	Weight int    `json:"weight"`
}

type ExternalServerYaml struct {
	Addr   string `yaml:"addr"`
	Weight int    `yaml:"weight"`
}

type ConfigJson struct {
	Amount                int                  `json:"amount"`
	Method                string               `json:"method"`
	Environment           string               `json:"environment"`
	Balanceer_port        int                  `json:"balancer_port"`
	Servers_port          int                  `json:"servers_port"`
	Health_check_interval int                  `json:"health_check_interval"`
	Servers               []ExternalServerJson `json:"servers"`
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func LoadConfig(path string) (*ConfigJson, error) {
	byteVal, err := readFile(path)
	if err != nil {
		log.Fatalf("Failed to load config from JSON file: %v", err)
		return &ConfigJson{}, err
	}
	var config ConfigJson
	err = json.Unmarshal(byteVal, &config)
	if err != nil {
		log.Fatalf("Failed to Unmarshal JSON: %v", err)
		return &ConfigJson{}, err
	}
	return &config, nil
}

func LoadServers(servers []ExternalServerJson) ([]*LbServer, error) {
	res := make([]*LbServer, 0, len(servers))
	for k, s := range servers {
		lb := NewLbServer(s.Addr, s.Weight)
		lb.name = "Server " + strconv.Itoa(k)
		res = append(res, lb)
	}
	return res, nil
}

func Loader(path string) ([]*LbServer, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml":
		res, err := ReadYaml(path)
		if err != nil {
			return []*LbServer{}, err
		}
		for _, s := range res {
			s.IsAlive()
		}
		return res, nil
	case ".json":
		res, err := ReadJson(path)
		if err != nil {
			return []*LbServer{}, err
		}
		for _, s := range res {
			s.IsAlive()
		}
		return res, nil
	default:
		log.Panic("No file provided")
		return []*LbServer{}, fmt.Errorf("No file provided")
	}
}

func ReadJson(path string) ([]*LbServer, error) {
	byteVal, err := readFile(path)
	if err != nil {
		return nil, err
	}

	var servers []ExternalServerJson
	if err := json.Unmarshal(byteVal, &servers); err != nil {
		return nil, err
	}

	res := make([]*LbServer, len(servers))
	for k, s := range servers {
		lbServer := NewLbServer(s.Addr, s.Weight)
		lbServer.name = strconv.Itoa(k)
		res[k] = lbServer
	}

	return res, nil
}

func ReadYaml(path string) ([]*LbServer, error) {
	byteVal, err := readFile(path)
	if err != nil {
		return nil, err
	}

	var servers []ExternalServerYaml
	if err := yaml.Unmarshal(byteVal, &servers); err != nil {
		return nil, err
	}

	res := make([]*LbServer, len(servers))
	for k, s := range servers {
		lbServer := NewLbServer(s.Addr, s.Weight)
		lbServer.name = strconv.Itoa(k)
		res = append(res, lbServer)
	}

	return res, nil
}
