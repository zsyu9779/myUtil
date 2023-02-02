package gopool

const (
	defaultScalaThreshold = 1
)

// Config is used to config pool.
type Config struct {
	// threshold for scale.
	// new goroutine is created if len(task chan) > ScaleThreshold.
	// defaults to defaultScalaThreshold.
	//扩容规模阈值，若len(task chan) > ScaleThreshold，则新建goroutine
	ScaleThreshold int32
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	c := &Config{
		ScaleThreshold: defaultScalaThreshold,
	}
	return c
}
