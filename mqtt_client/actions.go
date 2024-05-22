package mqtt_client

import (
	"errors"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Pub(pubClient mqtt.Client, topic string, qos byte, playload string) error {
	if ExitFlag {
		return errors.New("pub client exited")
	}
	if !pubClient.IsConnectionOpen() {
		return errors.New("pub client is not connected")
	}
	token := pubClient.Publish(topic, qos, false, playload)
	if token.Error() != nil {
		return fmt.Errorf("pub message to topic %s error:%s \n", topic, token.Error())
	}
	//return fmt.Errorf("pub %s to topic [%s]\n", playload, topic)
	return nil
}

func Sub(subClient mqtt.Client, qos byte, topic string, callback mqtt.MessageHandler) error {
	if ExitFlag {
		return errors.New("sub client exited")
	}
	token := subClient.Subscribe(topic, qos, callback)
	if token.WaitTimeout(time.Second) && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func PubTest(client mqtt.Client, topic string) {
	pubClient := client
	i := 1
	for !ExitFlag {
		payload := fmt.Sprintf("%d", i)
		err := Pub(pubClient, topic, 0, payload)
		if err != nil {
			log.Printf("pub message to topic %s error:%s \n", topic, err)
		} else {
			log.Printf("pub %s to topic [%s]\n", payload, topic)
		}
		i += 1
		time.Sleep(1 * time.Second)
	}
}

func SubTest(client mqtt.Client, topic string) {
	subClient := client
	for !ExitFlag {
		err := Sub(subClient, 0, topic, func(subClient mqtt.Client, msg mqtt.Message) {
			log.Printf("sub [%s] %s\n", msg.Topic(), string(msg.Payload()))
		})
		if err != nil {
			log.Printf("sub message to topic %s error:%s \n", topic, err)
		}
		time.Sleep(1 * time.Second)
	}
}

func PubSubTest(client mqtt.Client, topic string) {
	go SubTest(client, topic)
	PubTest(client, topic)
}
