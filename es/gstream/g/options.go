package g

type Options struct {
	ProjectId string
	TopicId   string
}

func NewOptions() *Options {
	return &Options{
		ProjectId: "",
		TopicId:   "EventStore",
	}
}

type OptionFunc func(*Options)

func WithProjectId(projectId string) OptionFunc {
	return func(o *Options) {
		o.ProjectId = projectId
	}
}

func WithTopicId(topicId string) OptionFunc {
	return func(o *Options) {
		o.TopicId = topicId
	}
}
