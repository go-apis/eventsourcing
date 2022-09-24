package db

import (
	"fmt"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Debug    bool
}

func (o *Config) DSN() string {
	var parts []string

	if o.Host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", o.Host))
	}

	if o.Port != 0 {
		parts = append(parts, fmt.Sprintf("port=%d", o.Port))
	}

	if o.Name != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", o.Name))
	}

	if o.User != "" {
		parts = append(parts, fmt.Sprintf("user=%s", o.User))
	}

	if o.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", o.Password))
	}

	parts = append(parts, "sslmode=disable")

	return strings.Join(parts, " ")
}

func NewConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 5432,
	}
}
