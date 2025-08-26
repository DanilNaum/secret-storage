package pg

import (
	"fmt"
)

type connConfig struct {
	host      string
	port      string
	username  string
	password  string
	dbName    string
	dncString string
}

// NewConnConfig creates a new connConfig instance with the provided connection parameters.
// It allows creating a connection configuration by specifying host, port, username, password, and database name.
func NewConnConfig(host, port, username, password, dbName string) *connConfig {
	return &connConfig{
		host:     host,
		port:     port,
		username: username,
		password: password,
		dbName:   dbName,
	}
}

// NewConnConfigFromDsnString creates a new connConfig instance using a pre-existing DSN connection string.
// It allows creating a connection configuration directly from a fully formed data source name string.
func NewConnConfigFromDsnString(dnsString string) *connConfig {
	return &connConfig{
		dncString: dnsString,
	}
}

func (c *connConfig) getDsn() string {
	if c.dncString != "" {
		return c.dncString
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.host,
		c.port,
		c.username,
		c.password,
		c.dbName,
	)
}
