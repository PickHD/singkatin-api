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
		RabbitMQ *RabbitMQ
		Tracer   *Tracer
	}

	Common struct {
		GrpcPort int
	}

	Server struct {
		AppPort int
		AppEnv  string
		AppName string
		AppID   string
	}

	Database struct {
		Port                 int
		Host                 string
		Name                 string
		UsersCollection      string
		ShortenersCollection string
		ClicksCollection     string
	}

	Redis struct {
		Host string
		Port int
		TTL  int
	}

	RabbitMQ struct {
		ConnURL              string
		QueueCreateShortener string
		QueueUpdateVisitor   string
		QueueUpdateShortener string
		QueueDeleteShortener string
	}

	Tracer struct {
		JaegerURL string
	}
)

func loadConfiguration() *Config {
	return &Config{
		Common: &Common{
			GrpcPort: getEnvInt("GRPC_PORT"),
		},
		Server: &Server{
			AppPort: getEnvInt("APP_PORT"),
			AppEnv:  getEnv("APP_ENV"),
			AppName: getEnv("APP_NAME"),
			AppID:   getEnv("APP_ID"),
		},
		Database: &Database{
			Port:                 getEnvInt("DB_PORT"),
			Host:                 getEnv("DB_HOST"),
			Name:                 getEnv("DB_NAME"),
			UsersCollection:      getEnv("DB_COLLECTION_USERS"),
			ShortenersCollection: getEnv("DB_COLLECTION_SHORTENERS"),
			ClicksCollection:     getEnv("DB_COLLECTION_CLICKS"),
		},
		Redis: &Redis{
			Host: getEnv("REDIS_HOST"),
			Port: getEnvInt("REDIS_PORT"),
			TTL:  getEnvInt("REDIS_TTL"),
		},
		RabbitMQ: &RabbitMQ{
			ConnURL:              getEnv("AMQP_SERVER_URL"),
			QueueCreateShortener: getEnv("AMQP_QUEUE_CREATE_SHORTENER"),
			QueueUpdateVisitor:   getEnv("AMQP_QUEUE_UPDATE_VISITOR_COUNT"),
			QueueUpdateShortener: getEnv("AMQP_QUEUE_UPDATE_SHORTENER"),
			QueueDeleteShortener: getEnv("AMQP_QUEUE_DELETE_SHORTENER"),
		},
		Tracer: &Tracer{
			JaegerURL: getEnv("JAEGER_URL"),
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
