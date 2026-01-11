package config

import "github.com/PickHD/singkatin-revamp/auth/internal/helper"

type (
	Configuration struct {
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

func loadConfiguration() *Configuration {
	return &Configuration{
		Common: &Common{
			JWTExpire: helper.GetEnvInt("JWT_EXPIRE"),
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
		Secret: &Secret{
			JWTSecret: helper.GetEnvString("JWT_SECRET"),
		},
		Tracer: &Tracer{
			JaegerURL: helper.GetEnvString("JAEGER_URL"),
		},
		Mailer: &Mailer{
			Host:     helper.GetEnvString("SMTP_HOST"),
			Port:     helper.GetEnvInt("SMTP_PORT"),
			Username: helper.GetEnvString("SMTP_USERNAME"),
			Password: helper.GetEnvString("SMTP_PASSWORD"),
			Sender:   helper.GetEnvString("SMTP_SENDER"),
			SSL:      helper.GetEnvInt("SMTP_SSL"),
			IsTLS:    helper.GetEnvBool("SMTP_IS_TLS"),
		},
	}
}

func NewConfig() *Configuration {
	return loadConfiguration()
}
