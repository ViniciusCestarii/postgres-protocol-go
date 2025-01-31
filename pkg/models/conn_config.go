package models

type ConnConfig struct {
	// Connection
	Port     int
	Hostname string
	Username string
	Database *string
	Password *string

	// Options
	Verbose *bool
}
