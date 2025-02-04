package models

type ConnConfig struct {
	Port     int
	Host     string
	User     string
	Database *string
	Password *string
}

type DriveConfig struct {
	Verbose *bool
}
