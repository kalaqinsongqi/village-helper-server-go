package main

import (
	"fmt"
	"os"
)

type Config struct {
	ProjectName string
	Version     string
	Description string
	DBPath      string
	JWTSecret   string
	JWTHours    int
	Port        string
}

var cfg = Config{
	ProjectName: "Village Helper Server",
	Version:     "0.1.0",
	Description: "village-helper Go后端服务",
	DBPath:      "./data/app.db",
	JWTSecret:   "your-super-secret-key-change-in-production",
	JWTHours:    24,
	Port:        "8000",
}

func ensureDataDir() {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.MkdirAll("data", 0755)
	}
}

func getDSN() string {
	dbPath := cfg.DBPath
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	return fmt.Sprintf("%s?_fk=1", dbPath)
}
