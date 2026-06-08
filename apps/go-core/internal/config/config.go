// config/config.go
package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

type App struct {
	Port         string   `env:"APP_PORT,required"`
	Env          string   `env:"APP_ENV" envDefault:"production"`
	FrontendURL  string   `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
	AllowOrigins []string `env:"ALLOW_ORIGINS"   envDefault:"http://localhost:3000"`
}

type DB struct {
	URL          string        `env:"DATABASE_URL,required"`
	MaxOpenConns int           `env:"DB_MAX_OPEN_CONNS"     envDefault:"25"`
	MaxIdleConns int           `env:"DB_MAX_IDLE_CONNS"     envDefault:"10"`
	ConnLifetime time.Duration `env:"DB_CONN_LIFETIME"      envDefault:"30m"`
	ConnIdleTime time.Duration `env:"DB_CONN_IDLE_TIME"     envDefault:"5m"`
}

type ImageKit struct {
	PublicKey   string `env:"IMAGEKIT_PUBLIC_KEY,required"`
	PrivateKey  string `env:"IMAGEKIT_PRIVATE_KEY,required"`
	UrlEndpoint string `env:"IMAGEKIT_URL_ENDPOINT,required"`
}
type Session struct {
	Secret       string `env:"SESSION_SECRET,required"`
	CookieName   string `env:"COOKIE_NAME" envDefault:"_zovly_secure_token"`
	CookieMaxAge int    `env:"COOKIE_MAX_AGE" envDefault:"604800"`
}

type OAuthConfig struct {
	ClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	ClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	RedirectURL  string `env:"GOOGLE_REDIRECT_URL,required"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR,required"`     // localhost:6379
	Password string `env:"REDIS_PASSWORD"`          // optional
	DB       int    `env:"REDIS_DB" envDefault:"0"` // default database
}

// type ResendConfig struct {
// 	APIKey    string `env:"RESEND_API_KEY,required"`
// 	FromEmail string `env:"RESEND_FROM_EMAIL,required"`
// 	FromName  string `env:"RESEND_FROM_NAME" envDefault:"Tixort"`
// }

// type AsynqConfig struct {
// 	Concurrency int `env:"ASYNQ_CONCURRENCY" envDefault:"10"`
// }

type PaddleConfig struct {
	APISecret      string `env:"PADDLE_SECRET_KEY,required"`
	WebhookSecret  string `env:"PADDLE_WEBHOOK_SECRET,required"`
	ClientSecret string `env:"PADDLE_CLIENT_SECRET,required"`
	}
type Config struct {
	App      App
	DB       DB
	Session  Session
	OAuth    OAuthConfig
	ImageKit ImageKit
	Paddle   PaddleConfig
	Redis RedisConfig
	// Resend   ResendConfig
	// Asynq    AsynqConfig
}

func MustLoad() *Config {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to load env %v", err)
	}
	return cfg
}
