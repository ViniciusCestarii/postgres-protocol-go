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
	PGSECURE := parseBoolEnv(os.Getenv("PGSECURE"))

	secure := "disable"
	if PGSECURE {
		secure = "require"
	}

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", PGHOST, PGUSER, PGPASSWORD, PGDATABASE, PGPORT, secure)

	PGURL := os.Getenv("PGURL")

	if PGURL != "" {
		connStr = PGURL
	}

	PGVERBOSE := parseBoolEnv(os.Getenv("PGVERBOSE"))

	driveConfig := models.DriveConfig{
		Verbose: PGVERBOSE,
	}

	return connStr, driveConfig
}

func parseBoolEnv(env string) bool {
	return env == "true"
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
