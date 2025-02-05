package models

type ConnConfig struct {
	Port     int
	Host     string
	User     string
	Secure   bool
	Database *string
	Password *string
}

type DriveConfig struct {
	Verbose bool
}
