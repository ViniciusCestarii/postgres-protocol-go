# postgres protocol go

## Folder Structure

```bash
postgres-protocol-go/
│── cmd/
│   ├── server/          # Main entry point for the server
│   ├── client/          # Optional: Client implementation for testing
│── internal/
│   ├── network/         # TCP server, connection management
│   ├── protocol/        # PostgreSQL wire protocol handling
│   ├── parser/          # SQL parser, message serialization/deserialization
│   ├── executor/        # Command execution logic
│   ├── storage/         # Storage engine, indexing, persistence
│   ├── auth/            # Authentication and authorization handling
│   ├── config/          # Configuration management (YAML, ENV, etc.)
│── pkg/
│   ├── utils/           # Shared utilities (logging, errors, helpers)
│   ├── models/          # Data structures for queries, results, etc.
│── tests/               # Integration and unit tests
│── go.mod               # Go module file
│── README.md            # Project documentation
```

## Aknowledgements

(Official Protocol Documentation)[https://www.postgresql.org/docs/current/protocol.html]
(Message Formats)[https://www.postgresql.org/docs/current/protocol-message-formats.html]