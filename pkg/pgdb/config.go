package pgdb

import (
	"fmt"
	"net/url"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SslMode  string
	Options  string

	Debug bool
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

	if o.SslMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", o.SslMode))
	} else {
		parts = append(parts, "sslmode=disable")
	}

	if o.Options != "" {
		parts = append(parts, fmt.Sprintf("options=%s", o.Options))
	}

	return strings.Join(parts, " ")
}

func (o *Config) URL() string {
	sslmode := "disable"
	if o.SslMode != "" {
		sslmode = o.SslMode
	}

	query := url.Values{
		"sslmode": []string{sslmode},
	}

	if o.Options != "" {
		query["options"] = []string{o.Options}
	}

	out := url.URL{
		Scheme:   "postgresql",
		Host:     fmt.Sprintf("%s:%d", o.Host, o.Port),
		Path:     o.Name,
		User:     url.UserPassword(o.User, o.Password),
		RawQuery: query.Encode(),
	}

	return out.String()
}

func NewConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 5432,
	}
}
