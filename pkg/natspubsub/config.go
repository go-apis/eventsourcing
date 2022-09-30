package natspubsub

type Config struct {
	Url     string
	Subject string
}

func NewConfig() *Config {
	return &Config{
		Url:     "nats://localhost:4222",
		Subject: "events",
	}
}
