package mqtt_client

import "log"

//这里举例说明了几种场景

func MQTTConnection(config Config) {
	client := connectByMQTT(config)
	action := config.Action
	switch action {
	case "pub":
		PubTest(client, config.Topic)
	case "sub":
		SubTest(client, config.Topic)
	case "pubsub":
		PubSubTest(client, config.Topic)
	default:
		log.Fatalf("Unsupported action: %s", action)
	}
}

func MQTTSConnection(config Config) {
	client := connectByMQTTS(config)
	action := config.Action
	switch action {
	case "pub":
		PubTest(client, config.Topic)
	case "sub":
		SubTest(client, config.Topic)
	case "pubsub":
		PubSubTest(client, config.Topic)
	default:
		log.Fatalf("Unsupported action: %s", action)
	}
}

func WSConnection(config Config) {
	client := connectByWS(config)
	action := config.Action
	switch action {
	case "pub":
		PubTest(client, config.Topic)
	case "sub":
		SubTest(client, config.Topic)
	case "pubsub":
		PubSubTest(client, config.Topic)
	default:
		log.Fatalf("Unsupported action: %s", action)
	}
}

func WSSConnection(config Config) {
	client := connectByWSS(config)
	action := config.Action
	switch action {
	case "pub":
		PubTest(client, config.Topic)
	case "sub":
		SubTest(client, config.Topic)
	case "pubsub":
		PubSubTest(client, config.Topic)
	default:
		log.Fatalf("Unsupported action: %s", action)
	}
}
