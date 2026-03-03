package config

import (
	"os"
	"strconv"
)

type (
	Config struct {
		Server   *Server
		Common   *Common
		RabbitMQ *RabbitMQ
		Tracer   *Tracer
		MinIO    *MinIO
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

	RabbitMQ struct {
		ConnURL           string
		QueueUploadAvatar string
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
		RabbitMQ: &RabbitMQ{
			ConnURL:           getEnv("AMQP_SERVER_URL"),
			QueueUploadAvatar: getEnv("AMQP_QUEUE_UPLOAD_AVATAR"),
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
