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

type JWTConfig struct {
	Secret     string
	ExpiryHour int
}

type MinioConfig struct {
	SecretAccessKey string
	UseSSL bool
	Endpoint string
	AccessKeyID string
}

type StorageConfig struct {
	MinioConfig MinioConfig
}

type Config struct{
	Database DatabaseConfig
	App ApplicationConfig
	SMTP EmailConfig
	JWT      JWTConfig
	Storage StorageConfig
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
		JWT: JWTConfig{
	    	Secret:     os.Getenv("JWT_SECRET"),
	    	ExpiryHour: 24,
		},
		Storage : StorageConfig {
			MinioConfig : MinioConfig{
				Endpoint: "127.0.0.1:9000",
				AccessKeyID: os.Getenv("MINIO_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("MINIO_SECRET_ACCESS_KEY"),
				UseSSL: false,
			},
		},

	}
}