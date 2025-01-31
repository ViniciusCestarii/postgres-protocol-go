package main

import (
	"fmt"
	"net"
	"postgres-protocol-go/internal/protocol"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:5432")
	if err != nil {
		fmt.Println("Error connecting to PostgreSQL:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to PostgreSQL server")

	handleConnection(conn)
}

func handleConnection(conn net.Conn) {
	dbName := "sagep-auth-development"
	password := "123456"
	connConfig := models.ConnConfig{
		Username: "postgres",
		Database: &dbName,
		Password: &password,
	}

	protocol.ProcessStartup(conn, connConfig)

	answer := make([]byte, 1024)

	_, err := conn.Read(answer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	utils.LogServerAnswer(answer)

	identifier := string(answer[0:1])

	switch identifier {
	case "R":
		err := protocol.ProcessAuth(conn, answer, connConfig)

		if err != nil {
			fmt.Println(err)
		}
	default:
		fmt.Println("Unknown message identifier:", identifier)
	}
}
