package gcppubsub

type Config struct {
	ProjectId string
	TopicId   string
}

func NewConfig() *Config {
	return &Config{
		ProjectId: "",
		TopicId:   "EventStore",
	}
}
