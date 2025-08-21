package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	defaultGRPCPort    = 9090
	defaultJWTDuration = 30 * time.Minute
)

type config struct {
	DBHost *string `env:"PSQL_DBHOST"`
	DBPort *string `env:"PSQL_DBPORT"`
	DBUser *string `env:"PSQL_DBUSER"`
	DBName *string `env:"PSQL_DBNAME"`
	DBPass *string `env:"PSQL_DBPASS"`

	JWTDurationSec *int    `env:"JWT_DURATION_SEC"`
	JWTSecret      *string `env:"JWT_SECRET"`

	GRPCPort *int `env:"GRPC_PORT"`
}

func NewConfig() (*config, error) {
	c := &config{}
	err := env.Parse(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetDSN returns the Data Source Name (DSN) from the database configuration.
// It returns an empty string if no DSN is set.
func (c *config) GetDSN() string {
	if c.DBHost == nil || c.DBPort == nil || c.DBUser == nil || c.DBName == nil || c.DBPass == nil {
		return ""
	}
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", *c.DBHost, *c.DBPort, *c.DBUser, *c.DBName, *c.DBPass)
}

func (c *config) GetGRPCPort() int {
	if c.GRPCPort == nil {
		return defaultGRPCPort
	}
	return *c.GRPCPort
}
func (c *config) GetJWTSecret() string {
	return *c.JWTSecret
}

func (c *config) GetJWTDuration() time.Duration {
	if c.JWTDurationSec == nil {
		return defaultJWTDuration
	}
	return time.Second * time.Duration(*c.JWTDurationSec)
}
