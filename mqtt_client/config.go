package mqtt_client

var ExitFlag = false

type Config struct {
	Host     string
	Port     int
	Action   string
	Topic    string
	Username string
	Password string
	Qos      int
	Tls      bool
	CaCert   string
}
