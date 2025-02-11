# Postgres Protocol Go

This project implements the PostgreSQL wire protocol in Go using only the standard library.

## Usage

```go
func main() {
	connStr := "postgres://postgres:123456@localhost:5432/postgres"

	driveConfig := models.DriveConfig{Verbose: true,}

	pgConnection, err := protocol.NewPgConnection(connStr, driveConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	userToFind := "postgres"

	res, err := pgConnection.Query("SELECT * FROM pg_user WHERE usename = $1;", userToFind)

	if err != nil {
		fmt.Println(err)
		pgConnection.Close()
		return
	}

	fmt.Println("Postgres user: ", res.Rows)

	pgConnection.Close()
}
```

## Features

- Flexible Connection Handling
	- Supports both URL-style connection strings (postgres://user:pass@host:port/db)
	- Supports key-value connection strings (`host=localhost port=5432`)
- SSL/TLS Support
	- Automatic SSL/TLS negotiation when `sslmode=require`
	- Secure encrypted connections
-	Query Interface
	-	Simple query protocol support
	- Extended query protocol with parameter binding
	- Support for parameterized queries using $1, $2 etc.
- Connection Configuration
	- Configurable verbose mode for debugging
	- Custom drive configuration options via models.DriveConfig
-	Clean Resource Management
	- Proper connection termination

## Getting Started

To run the server, use the following commands:

```bash
cp .env.example .env
```

```bash
go run cmd/server.go
```

## Folder Structure

```bash
postgres-protocol-go/
│── cmd/
│   ├── server/          # Main entry point for the server
│── internal/
│   ├── pool/            # Buff writer
│   ├── protocol/        # PostgreSQL wire protocol handling
│── pkg/
│   ├── utils/           # Shared utilities (logging, errors, helpers)
│   ├── models/          # Data structures for queries, results, etc.
│── tests/               # Integration and unit tests
│── go.mod               # Go module file
│── README.md            # Project documentation
```

## Testing

To run the tests, use the following commands:

```bash
go test ./tests/...
```

## Acknowledgements

[Official Protocol Documentation](https://www.postgresql.org/docs/current/protocol.html)

[Message Formats](https://www.postgresql.org/docs/current/protocol-message-formats.html)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
