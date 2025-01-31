package utils

import (
	"bufio"
	"os"
	"postgres-protocol-go/pkg/models"
	"strconv"
	"strings"
)

func GetEnvConnConfig() (models.ConnConfig, error) {
	loadEnvFile(".env")

	port, err := strconv.Atoi(*getEnv("PGPORT", "5432"))
	if err != nil {
		return models.ConnConfig{}, err
	}

	verbose, err := strconv.ParseBool(*getEnv("PGVERBOSE", "false"))
	if err != nil {
		verbose = false // Default to false if invalid
	}

	connConfig := models.ConnConfig{
		Port:     port,
		Username: *getEnv("PGUSER", "postgres"),
		Hostname: *getEnv("PGHOST", "localhost"),
		Database: getEnv("PGDATABASE", "postgres"),
		Password: getEnv("PGPASSWORD", ""),
		Verbose:  &verbose,
	}

	return connConfig, nil
}

func getEnv(key, defaultValue string) *string {
	if value, exists := os.LookupEnv(key); exists {
		return &value
	}
	return &defaultValue
}

// func to load variables from .env file
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue // Skip comments and empty lines
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Ignore malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set the environment variable only if not already set
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}
