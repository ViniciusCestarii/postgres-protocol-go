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

	value, err := pgConnection.Query("SELECT * FROM user;")

	if err != nil {
		fmt.Println(err)
		pgConnection.Close()
		return
	}

	fmt.Println("All user: ", value)

	value, err = pgConnection.Query("SELECT * FROM pg_stat_activity;")

	if err != nil {
		fmt.Println(err)
		pgConnection.Close()
		return
	}

	fmt.Println("All pg_stat_activity", value)

	pgConnection.Close()
}
