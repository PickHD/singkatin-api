package config

import (
	"github.com/PickHD/singkatin-revamp/upload/internal/helper"
)

type (
	Configuration struct {
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
		RabbitMQ: &RabbitMQ{
			ConnURL:           helper.GetEnvString("AMQP_SERVER_URL"),
			QueueUploadAvatar: helper.GetEnvString("AMQP_QUEUE_UPLOAD_AVATAR"),
		},
		Tracer: &Tracer{
			JaegerURL: helper.GetEnvString("JAEGER_URL"),
		},
		MinIO: &MinIO{
			Endpoint:  helper.GetEnvString("MINIO_ENDPOINT"),
			AccessKey: helper.GetEnvString("MINIO_ACCESSKEY"),
			SecretKey: helper.GetEnvString("MINIO_SECRETKEY"),
			Bucket:    helper.GetEnvString("MINIO_BUCKET"),
			UseSSL:    helper.GetEnvBool("MINIO_USE_SSL"),
			Location:  helper.GetEnvString("MINIO_LOCATION"),
		},
	}
}

func NewConfig() *Configuration {
	return loadConfiguration()
}
