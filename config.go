package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name         string   `yaml:"name"`
	Execpath     string   `yaml:"exec_path"`
	Path         string   `yaml:"path"`
	Args         []string `yaml:"args"`
	Restart      bool     `yaml:"restart"`
	Restartdelay int      `yaml:"restart_delay"`
	Errorlogs    string   `yaml:"error_logs"`
}

func readconfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error: %v", err)
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Printf("Error: %v", err)
		return Config{}, err
	}
	// log.Printf("Exec_path: %s, Path: %s", config.Execpath, config.Path)
	return config, nil
}
