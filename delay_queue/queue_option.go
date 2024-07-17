package queue

type Option func(*Options)

type Options struct {
	topic   string
	handler handleFunc
}

func WithTopic(topic string) Option {
	return func(opts *Options) {
		opts.topic = topic
	}
}

func WithHandler(handler handleFunc) Option {
	return func(opts *Options) {
		opts.handler = handler
	}
}
