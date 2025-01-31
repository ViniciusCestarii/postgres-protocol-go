package main

import (
	"fmt"
	"postgres-protocol-go/internal/protocol"
	"postgres-protocol-go/pkg/models"
)

func main() {
	dbName := "sagep-auth-development"
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

	_, err := protocol.NewPgConnection(connConfig, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
}
