# Postgres Protocol Go

This project implements the PostgreSQL wire protocol in Go using only the standard library.

## Getting Started

To run the server, use the following command:

```bash
go run cmd/server/main.go
```

## Folder Structure

```bash
postgres-protocol-go/
│── cmd/
│   ├── server/          # Main entry point for the server
│   ├── client/          # Optional: Client implementation for testing
│── internal/
│   ├── protocol/        # PostgreSQL messages constants
│   ├── protocol/        # PostgreSQL wire protocol handling
│── pkg/
│   ├── utils/           # Shared utilities (logging, errors, helpers)
│   ├── models/          # Data structures for queries, results, etc.
│── tests/               # Integration and unit tests
│── go.mod               # Go module file
│── README.md            # Project documentation
```

## Aknowledgements

[Official Protocol Documentation](https://www.postgresql.org/docs/current/protocol.html)
[Message Formats](https://www.postgresql.org/docs/current/protocol-message-formats.html)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
