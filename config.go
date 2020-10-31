package main

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port     int
	Env      string
	Pepper   string
	HMACKey  string
	Database PostgresConfig
	Mailgun  MailgunConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func (c Config) ConnectionInfo() string {
	if c.Database.Password == "" {
		return fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s sslmode=disable",
			c.Database.Host, c.Database.Port, c.Database.User, c.Database.Name,
		)
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name,
	)
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port:    3000,
		Env:     "dev",
		Pepper:  "secret-random-string",
		HMACKey: "secret-hmac-key",
		Database: PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Name:     "postgres",
		},
	}
}

type MailgunConfig struct {
	APIKey       string
	PublicAPIKey string
	Domain       string
}

func LoadConfig() Config {
	var c Config
	c = DefaultConfig()
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	return c
}
