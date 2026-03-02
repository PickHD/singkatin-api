package config

import (
	"os"
	"strconv"
)

type (
	Config struct {
		Server      *Server
		Common      *Common
		Database    *Database
		RabbitMQ    *RabbitMQ
		Secret      *Secret
		Tracer      *Tracer
		MinIO       *MinIO
		HttpService *HttpService
	}

	Common struct {
		JWTExpire int
		GRPCPort  string
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

	RabbitMQ struct {
		ConnURL              string
		QueueCreateShortener string
		QueueUpdateVisitor   string
		QueueUploadAvatar    string
		QueueUpdateShortener string
		QueueDeleteShortener string
	}

	Secret struct {
		JWTSecret string
	}

	Tracer struct {
		JaegerURL string
	}

	MinIO struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
		Location  string
	}

	HttpService struct {
		ShortenerBaseAPIURL string
	}
)

func loadConfiguration() *Config {
	return &Config{
		Common: &Common{
			JWTExpire: getEnvInt("JWT_EXPIRE"),
			GRPCPort:  getEnv("GRPC_SHORTENER_HOST"),
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
		},
		RabbitMQ: &RabbitMQ{
			ConnURL:              getEnv("AMQP_SERVER_URL"),
			QueueCreateShortener: getEnv("AMQP_QUEUE_CREATE_SHORTENER"),
			QueueUpdateVisitor:   getEnv("AMQP_QUEUE_UPDATE_VISITOR"),
			QueueUploadAvatar:    getEnv("AMQP_QUEUE_UPLOAD_AVATAR"),
			QueueUpdateShortener: getEnv("AMQP_QUEUE_UPDATE_SHORTENER"),
			QueueDeleteShortener: getEnv("AMQP_QUEUE_DELETE_SHORTENER"),
		},
		Secret: &Secret{
			JWTSecret: getEnv("JWT_SECRET"),
		},
		Tracer: &Tracer{
			JaegerURL: getEnv("JAEGER_URL"),
		},
		MinIO: &MinIO{
			Endpoint:  getEnv("MINIO_ENDPOINT"),
			AccessKey: getEnv("MINIO_ACCESSKEY"),
			SecretKey: getEnv("MINIO_SECRETKEY"),
			Bucket:    getEnv("MINIO_BUCKET"),
			UseSSL:    getEnvBool("MINIO_USE_SSL"),
			Location:  getEnv("MINIO_LOCATION"),
		},
		HttpService: &HttpService{
			ShortenerBaseAPIURL: getEnv("SHORTENER_BASE_API_URL"),
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
	eBool, _ := strconv.ParseBool(os.Getenv(e))

	return eBool
}
