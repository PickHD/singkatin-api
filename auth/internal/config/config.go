package config

import (
	"os"
	"strconv"
)

type (
	Config struct {
		Server   *Server
		Common   *Common
		Database *Database
		Redis    *Redis
		Secret   *Secret
		Tracer   *Tracer
		Mailer   *Mailer
	}

	Common struct {
		JWTExpire int
	}

	Server struct {
		AppPort int
		AppEnv  string
		AppName string
		AppID   string
		BaseURL string
	}

	Database struct {
		Port                 int
		Host                 string
		Name                 string
		UsersCollection      string
		ShortenersCollection string
	}

	Redis struct {
		Host string
		Port int
		TTL  int
	}

	Secret struct {
		JWTSecret string
	}

	Tracer struct {
		JaegerURL string
	}

	Mailer struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
		IsTLS    bool
		SSL      int
	}
)

func loadConfiguration() *Config {
	return &Config{
		Common: &Common{
			JWTExpire: getEnvInt("JWT_EXPIRE"),
		},
		Server: &Server{
			AppPort: getEnvInt("APP_PORT"),
			AppEnv:  getEnv("APP_ENV"),
			AppName: getEnv("APP_NAME"),
			AppID:   getEnv("APP_ID"),
			BaseURL: getEnv("APP_BASE_URL"),
		},
		Database: &Database{
			Port:                 getEnvInt("DB_PORT"),
			Host:                 getEnv("DB_HOST"),
			Name:                 getEnv("DB_NAME"),
			UsersCollection:      getEnv("DB_COLLECTION_USERS"),
			ShortenersCollection: getEnv("DB_COLLECTION_SHORTENERS"),
		},
		Redis: &Redis{
			Host: getEnv("REDIS_HOST"),
			Port: getEnvInt("REDIS_PORT"),
			TTL:  getEnvInt("REDIS_TTL"),
		},
		Secret: &Secret{
			JWTSecret: getEnv("JWT_SECRET"),
		},
		Tracer: &Tracer{
			JaegerURL: getEnv("JAEGER_URL"),
		},
		Mailer: &Mailer{
			Host:     getEnv("SMTP_HOST"),
			Port:     getEnvInt("SMTP_PORT"),
			Username: getEnv("SMTP_USERNAME"),
			Password: getEnv("SMTP_PASSWORD"),
			Sender:   getEnv("SMTP_SENDER"),
			SSL:      getEnvInt("SMTP_SSL"),
			IsTLS:    getEnvBool("SMTP_IS_TLS"),
		},
	}
}

func Load() *Config {
	return loadConfiguration()
}

func getEnv(e string) string {
	return os.Getenv(e)
}

func getEnvInt(e string) int {
	eInt, _ := strconv.Atoi(os.Getenv(e))

	return eInt
}

func getEnvBool(e string) bool {
	eBoolean, _ := strconv.ParseBool(e)

	return eBoolean
}
