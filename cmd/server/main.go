package main

import (
	"fmt"
	"postgres-protocol-go/internal/protocol"
	"postgres-protocol-go/pkg/utils"
)

func main() {
	connConfig, err := utils.GetEnvConnConfig()

	if err != nil {
		fmt.Println(err)
		return
	}

	pgConnection, err := protocol.NewPgConnection(connConfig, nil)

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
