package local

import (
	"fmt"
	"strings"
)

type Options struct {
	Host     string
	Port     int
	DbName   string
	User     string
	Password string
	Debug    bool
}

func (o *Options) DSN() string {
	var parts []string

	if o.Host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", o.Host))
	}

	if o.Port != 0 {
		parts = append(parts, fmt.Sprintf("port=%d", o.Port))
	}

	if o.DbName != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", o.DbName))
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

func NewOptions() *Options {
	return &Options{
		Host: "localhost",
		Port: 5432,
	}
}

type OptionFunc func(*Options)

func WithDbHost(host string) OptionFunc {
	return func(o *Options) {
		o.Host = host
	}
}

func WithDbPort(port int) OptionFunc {
	return func(o *Options) {
		o.Port = port
	}
}

func WithDbName(dbName string) OptionFunc {
	return func(o *Options) {
		o.DbName = dbName
	}
}

func WithDbUser(user string) OptionFunc {
	return func(o *Options) {
		o.User = user
	}
}

func WithDbPassword(password string) OptionFunc {
	return func(o *Options) {
		o.Password = password
	}
}

func WithDebug(debug bool) OptionFunc {
	return func(o *Options) {
		o.Debug = debug
	}
}
