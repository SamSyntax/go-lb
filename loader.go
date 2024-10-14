package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ExternalServerJson struct {
	Addr   string `json:"addr"`
	Weight int    `json:"weight"`
}

type ExternalServerYaml struct {
	Addr   string `yaml:"addr"`
	Weight int    `yaml:"weight"`
}

func Loader(path string) []LbServer {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml":
		res := ReadYaml(path)
		return res
	case ".json":
		res := ReadJson(path)
		return res
	default:
		fmt.Println("No file provided")
		return []LbServer{}
	}
}

func ReadJson(path string) []LbServer {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to load servers from JSON file %v", err)
	}
	defer file.Close()

	byteVal, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var servers []ExternalServerJson
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
func ReadYaml(path string) []LbServer {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to load servers from JSON file %v", err)
	}
	defer file.Close()

	byteVal, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var servers []ExternalServerYaml
	err = yaml.Unmarshal(byteVal, &servers)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	res := make([]LbServer, 0, len(servers))
	for k, s := range servers {
		lbServer := NewLbServer(s.Addr, s.Weight)
		lbServer.name = strconv.Itoa(k)
		res = append(res, *lbServer)
	}

	return res
}
