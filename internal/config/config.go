package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"development"`
	Port   string `env:"PORT" envDefault:"8080"`
	DB     DBConfig
	Redis  RedisConfig
}

type DBConfig struct {
	DSN             string        `env:"DATABASE_DSN" envDefault:"host=localhost user=postgres password=postgres dbname=users port=5432 sslmode=disable TimeZone=UTC"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"30m"`
}

type RedisConfig struct {
	Addr    string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	UserTTL time.Duration `env:"REDIS_USER_TTL" envDefault:"30s"`
}

func Load() (*Config, error) {
	// Load .env file if it exists, but don't fail if it doesn't
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}
