package main

import (
	"fmt"
	"postgres-protocol-go/internal/protocol"
	"postgres-protocol-go/pkg/models"
)

func main() {
	dbName := "postgres"
	password := "123456"
	vebose := true
	connConfig := models.ConnConfig{
		Port:     5432,
		Username: "postgres",
		Hostname: "localhost",
		Database: &dbName,
		Password: &password,
		Verbose:  &vebose,
	}

	pgConnection, err := protocol.NewPgConnection(connConfig, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	value, err := pgConnection.Query("SELECT * FROM user;")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(value)
}
