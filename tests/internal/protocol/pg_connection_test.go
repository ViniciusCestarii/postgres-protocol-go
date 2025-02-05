package protocol_test

import (
	"fmt"
	"net"
	"postgres-protocol-go/internal/protocol"
	"postgres-protocol-go/pkg/models"
	"testing"
)

// todo: improve mock server to handle authentication
func startMockPostgresServer(t *testing.T, authSuccess bool) (string, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}

	addr := listener.Addr().String()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()

			buffer := make([]byte, 1024)
			n, _ := conn.Read(buffer)

			if authSuccess {
				conn.Write([]byte("R\x00\x00\x00\x08\x00\x00\x00\x00")) // Auth OK response
				conn.Write([]byte("Z\x00\x00\x00\x05I"))                // ReadyForQuery response
			} else {
				conn.Write([]byte("E\x00\x00\x00\x12Invalid password")) // Error response
			}

			conn.Read(buffer[:n])
		}
	}()

	return addr, func() { listener.Close() }
}

func TestNewPgConnection(t *testing.T) {
	tests := []struct {
		name        string
		authSuccess bool
		expectErr   bool
	}{
		{
			name:        "Valid PostgreSQL Connection with Authentication",
			authSuccess: true,
			expectErr:   false,
		},
		{
			name:        "Invalid Authentication",
			authSuccess: false,
			expectErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println("Running test: ", tc.name)
			mockAddr, cleanup := startMockPostgresServer(t, tc.authSuccess)
			defer cleanup()

			connURL := fmt.Sprintf("postgres://postgres:123456@%s/postgres", mockAddr)
			driveConfig := models.DriveConfig{
				Verbose: true,
			}

			conn, err := protocol.NewPgConnection(connURL, driveConfig)

			if (err != nil) != tc.expectErr {
				t.Fatalf("Test %q failed: expected error = %v, got %v", tc.name, tc.expectErr, err)
			}

			if conn != nil {
				defer conn.Close()
			}
		})
	}
}
