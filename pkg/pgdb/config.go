package pgdb

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
)

type Config struct {
	Host      string
	Port      int
	User      string
	Password  string
	Name      string
	SslMode   string
	Options   string
	AwsRegion string

	Debug bool
}

func (o *Config) DSN(ctx context.Context) (string, error) {
	var parts []string

	dbPass := o.Password
	dbEndpoint := fmt.Sprintf("%s:%d", o.Host, o.Port)

	// if the password doesn't exist lets try using AWS directly
	if len(o.Password) == 0 && len(o.AwsRegion) > 0 {
		awscfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return "", err
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, dbEndpoint, o.AwsRegion, o.User, awscfg.Credentials)
		if err != nil {
			return "", err
		}

		// set the pass
		dbPass = authenticationToken
	}

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
		parts = append(parts, fmt.Sprintf("password=%s", dbPass))
	}

	if o.SslMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", o.SslMode))
	}

	if o.Options != "" {
		parts = append(parts, fmt.Sprintf("options=%s", o.Options))
	}

	return strings.Join(parts, " "), nil
}

func (o *Config) URL(ctx context.Context) (string, error) {
	dbPass := o.Password
	dbEndpoint := fmt.Sprintf("%s:%d", o.Host, o.Port)

	// if the password doesn't exist lets try using AWS directly
	if len(o.Password) == 0 && len(o.AwsRegion) > 0 {
		awscfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return "", err
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, dbEndpoint, o.AwsRegion, o.User, awscfg.Credentials)
		if err != nil {
			return "", err
		}

		// set the pass
		dbPass = authenticationToken
	}

	query := url.Values{}
	if o.SslMode != "" {
		query["sslmode"] = []string{o.SslMode}
	}

	if o.Options != "" {
		query["options"] = []string{o.Options}
	}

	out := url.URL{
		Scheme:   "postgresql",
		Host:     dbEndpoint,
		Path:     o.Name,
		User:     url.UserPassword(o.User, dbPass),
		RawQuery: query.Encode(),
	}

	return out.String(), nil
}

func NewConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 5432,
	}
}
