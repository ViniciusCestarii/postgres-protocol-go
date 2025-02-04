package main

import (
	"bufio"
	"fmt"
	"os"
	"postgres-protocol-go/internal/protocol"
	"postgres-protocol-go/pkg/models"
	"strings"
)

func main() {
	connStr, pgConnectionConfig := getConfig()

	pgConnection, err := protocol.NewPgConnection(connStr, pgConnectionConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := pgConnection.Query("SELECT * FROM pg_user")

	if err != nil {
		fmt.Println(err)
		pgConnection.Close()
		return
	}

	fmt.Println("All users: ", res.Rows)

	userToFind := "postgres"

	res, err = pgConnection.Query("SELECT * FROM pg_user WHERE usename = $1;", userToFind)

	if err != nil {
		fmt.Println(err)
		pgConnection.Close()
		return
	}

	fmt.Println("Postgres user: ", res.Rows)

	pgConnection.Close()
}

func getConfig() (string, models.DriveConfig) {
	loadEnvFile(".env")

	PGHOST := os.Getenv("PGHOST")
	PGUSER := os.Getenv("PGUSER")
	PGPASSWORD := os.Getenv("PGPASSWORD")
	PGPORT := os.Getenv("PGPORT")
	PGDATABASE := os.Getenv("PGDATABASE")

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", PGHOST, PGUSER, PGPASSWORD, PGDATABASE, PGPORT)

	PGURL := os.Getenv("PGURL")

	if PGURL != "" {
		connStr = PGURL
	}

	PGVERBOSE := os.Getenv("PGVERBOSE")

	var verbose bool
	if PGVERBOSE == "true" {
		verbose = true
	} else if PGVERBOSE == "false" {
		verbose = false
	}

	driveConfig := models.DriveConfig{
		Verbose: verbose,
	}

	return connStr, driveConfig
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
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}
