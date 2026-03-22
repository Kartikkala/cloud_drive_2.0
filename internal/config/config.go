package config

import (
	"os"
)

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

type DatabaseConfig struct {
	User     string
	Password string
	Hostname string
	DBName   string
	Port     uint16
	Timezone string
}

type ApplicationConfig struct {
	RESTPort    uint16
	HostAddress string
}

type EmailConfig struct {
	Email    string
	Password string
	Port     uint16
}

type JWTConfig struct {
	Secret     string
	ExpiryHour int
}

type MinioConfig struct {
	SecretAccessKey string
	UseSSL          bool
	Endpoint        string
	AccessKeyID     string
}

type StorageConfig struct {
	MinioConfig   MinioConfig
	BucketName    string
	HLSBucketName string
}

type NATSConfig struct {
	URL string
}

type Config struct {
	Database DatabaseConfig
	App      ApplicationConfig
	SMTP     EmailConfig
	JWT      JWTConfig
	Storage  StorageConfig
	NATS     NATSConfig
}

func NewConfig() *Config {
	useSSL := false
	if getEnvOrDefault("MINIO_USE_SSL", "false") == "true" {
		useSSL = true
	}

	return &Config{
		Database: DatabaseConfig{
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASS"),
			DBName:   os.Getenv("POSTGRES_DBNAME"),
			Hostname: getEnvOrDefault("POSTGRES_HOSTNAME", "127.0.0.1"),
			Port:     uint16(5432),
			Timezone: getEnvOrDefault("POSTGRES_TIMEZONE", "Asia/Kolkata"),
		},
		App: ApplicationConfig{
			RESTPort:    8080,
			HostAddress: getEnvOrDefault("HOST_ADDRESS", "127.0.0.1"),
		},
		SMTP: EmailConfig{
			Email:    os.Getenv("EMAIL_ADDRESS"),
			Password: os.Getenv("EMAIL_PASSWORD"),
			Port:     uint16(587),
		},
		JWT: JWTConfig{
			Secret:     os.Getenv("JWT_SECRET"),
			ExpiryHour: 24,
		},
		Storage: StorageConfig{
			MinioConfig: MinioConfig{
				Endpoint:        getEnvOrDefault("MINIO_ENDPOINT", "127.0.0.1:9000"),
				AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("MINIO_SECRET_ACCESS_KEY"),
				UseSSL:          useSSL,
			},
			BucketName:    getEnvOrDefault("STORAGE_BUCKET_NAME", "cloud-drive"),
			HLSBucketName: getEnvOrDefault("STORAGE_HLS_BUCKET_NAME", "cloud-drive-hls"),
		},
		NATS: NATSConfig{
			URL: getEnvOrDefault("NATS_URL", "nats://127.0.0.1:4222"),
		},
	}
}