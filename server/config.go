package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kenaito-notebook-server/services"
)

type Config struct {
	Port    int
	BaseURL string
	SMTP    services.EmailConfig
}

func loadProperties(filename string) map[string]string {
	props := make(map[string]string)

	// Try executable directory first
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		props = loadFromFile(filepath.Join(exeDir, filename))
		if len(props) > 0 {
			return props
		}
	}

	// Try working directory
	props = loadFromFile(filename)
	return props
}

func loadFromFile(path string) map[string]string {
	props := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		return props
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			props[key] = value
		}
	}

	return props
}

func DefaultConfig() *Config {
	props := loadProperties("config.properties")

	port := 8080
	if p := props["server.port"]; p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	baseURL := props["server.base_url"]
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", port)
	}

	return &Config{
		Port:    port,
		BaseURL: baseURL,
		SMTP: services.EmailConfig{
			Host:     props["smtp.host"],
			Port:     props["smtp.port"],
			Username: props["smtp.username"],
			Password: props["smtp.password"],
			From:     props["smtp.from"],
		},
	}
}
