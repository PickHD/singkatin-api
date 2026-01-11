package config

import "github.com/PickHD/singkatin-revamp/shortener/internal/helper"

type (
	Configuration struct {
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

func loadConfiguration() *Configuration {
	return &Configuration{
		Common: &Common{
			GrpcPort: helper.GetEnvInt("GRPC_PORT"),
		},
		Server: &Server{
			AppPort: helper.GetEnvInt("APP_PORT"),
			AppEnv:  helper.GetEnvString("APP_ENV"),
			AppName: helper.GetEnvString("APP_NAME"),
			AppID:   helper.GetEnvString("APP_ID"),
		},
		Database: &Database{
			Port:                 helper.GetEnvInt("DB_PORT"),
			Host:                 helper.GetEnvString("DB_HOST"),
			Name:                 helper.GetEnvString("DB_NAME"),
			UsersCollection:      helper.GetEnvString("DB_COLLECTION_USERS"),
			ShortenersCollection: helper.GetEnvString("DB_COLLECTION_SHORTENERS"),
		},
		Redis: &Redis{
			Host: helper.GetEnvString("REDIS_HOST"),
			Port: helper.GetEnvInt("REDIS_PORT"),
			TTL:  helper.GetEnvInt("REDIS_TTL"),
		},
		RabbitMQ: &RabbitMQ{
			ConnURL:              helper.GetEnvString("AMQP_SERVER_URL"),
			QueueCreateShortener: helper.GetEnvString("AMQP_QUEUE_CREATE_SHORTENER"),
			QueueUpdateVisitor:   helper.GetEnvString("AMQP_QUEUE_UPDATE_VISITOR_COUNT"),
			QueueUpdateShortener: helper.GetEnvString("AMQP_QUEUE_UPDATE_SHORTENER"),
			QueueDeleteShortener: helper.GetEnvString("AMQP_QUEUE_DELETE_SHORTENER"),
		},
		Tracer: &Tracer{
			JaegerURL: helper.GetEnvString("JAEGER_URL"),
		},
	}
}

func NewConfig() *Configuration {
	return loadConfiguration()
}
