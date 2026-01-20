package config

import (
	"os"
)

type DatabaseConfig struct {
	User string
	Password string
	Hostname string
	DBName string
	Port uint16
	Timezone string
}

type ApplicationConfig struct {
	RESTPort uint16
	HostAddress string
}

type EmailConfig struct {
	Email string
	Password string
	Port uint16
}

type Config struct{
	Database DatabaseConfig
	App ApplicationConfig
	SMTP EmailConfig
}

func NewConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			User : os.Getenv("POSTGRES_USER"),
			Password : os.Getenv("POSTGRES_PASS"),
	 		DBName : os.Getenv("POSTGRES_DBNAME"),
			Hostname : "127.0.0.1",			
			Port : uint16(5432),
			Timezone : "Asia/Kolkata",
		},
		App: ApplicationConfig{
			RESTPort: 8080,
			HostAddress: "127.0.0.1",
		},
		SMTP: EmailConfig{
			Email: os.Getenv("EMAIL_ADDRESS"),
			Password: os.Getenv("EMAIl_PASSWORD"),
			Port: uint16(587),
		},
	}
}