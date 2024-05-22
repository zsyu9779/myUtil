package mqtt_client

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

var mqttClient mqtt.Client

const brokerHost = "emqt1.ahaent.cn:18883" //tcp
//const brokerHost = "emqt1.ahaent.cn:11883" //ws

func newMqttClient() error {
	opt := mqtt.NewClientOptions()
	opt.AddBroker(brokerHost)
	opt.SetClientID(uuid.New().String())
	opt.SetKeepAlive(5 * time.Second)
	opt.SetPingTimeout(15 * time.Second)
	opt.SetWriteTimeout(1 * time.Second)
	opt.SetCleanSession(true)
	opt.AutoReconnect = true
	opt.SetMaxReconnectInterval(5 * time.Second)
	opt.SetOnConnectHandler(func(_ mqtt.Client) {
		fmt.Printf("reconnect to mqtt broker %v \n", time.Now())
	})
	opt.OnConnectionLost = func(_ mqtt.Client, err error) {
		fmt.Printf("mqtt connection lost, err: %v  \n", err.Error())
	}
	opt.OnConnect = func(client mqtt.Client) {
		fmt.Printf("mqtt connection established \n")
		topicMap := make(map[string]byte)
		topicMap["/test/sub/topic"] = 0
		if token := mqttClient.SubscribeMultiple(topicMap, func(client mqtt.Client, message mqtt.Message) {
			fmt.Printf("SubscribeMultiple:%v\n", message.Topic())
		}); token.Wait() && token.Error() != nil {
			fmt.Printf("mqtt subscribe services reply err:%v\n", token.Error())
		}
	}
	mqttClient = mqtt.NewClient(opt)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt connection lost, err: %v", token.Error())
	}
	return nil
}

func TestMqttPubSub(t *testing.T) {
	err := newMqttClient()
	if err != nil {
		t.Error(err)
		return
	}
	var ops uint64
	for i := 0; i < 100; i++ {
		go func() {
			//共享订阅1000个主题
			for i := 0; i < 1000; i++ {
				time.Sleep(time.Millisecond * 300) // sleep 300ms
				//topic := fmt.Sprintf("$share/group/test/%s", uuid.New().String())
				topic := fmt.Sprintf("$share/g/test/%d", i)
				token := mqttClient.Subscribe(topic, 0, func(client mqtt.Client, message mqtt.Message) {
					log.Printf("client.Subscribe Topic: %s Payload: %s \n ", topic, message.Payload())
				})
				ack := token.WaitTimeout(2 * time.Second)
				if !ack {
					fmt.Printf("sub timeout:%d topic:%s \n", 3, topic)
				}
				atomic.AddUint64(&ops, 1)
			}
		}()

		go func() {
			for i := 0; i < 10000; i++ {
				time.Sleep(time.Millisecond * 20)
				body := `{"bid":"84399fd7-93f0-4724-b257-xxxxxxx","tid":"895c12bf-b71d-4dcd-xxxxxxxx","timestamp":1665578880925,"method":"flighttask_prepare","data":{"execute_time":1665583800000,"file":{"fingerprint":"xxxxxxxxxxxxxxxxxx","url":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},"xxxxx":"2af2e56f-xxxx-xxxx-xxxx-fb3cf0398412","task_type":1,"wayline_type":0}}`
				//topic := fmt.Sprintf("/test/%s", uuid.New().String())
				//topic := fmt.Sprintf("/test/%d", i%1000)
				topic := fmt.Sprintf("$share/g/test/%d", i%1000)

				if err := mqttClient.Publish(topic, 0, false, []byte(body)).Error(); err != nil {
					fmt.Printf("publish qos:0 topic:%s err:%v \n", topic, err)
				}
				token := mqttClient.Publish(topic, 0, false, []byte(body))
				if token.Error() != nil {
					fmt.Printf("publish qos:2 topic:%s err:%v \n", topic, token.Error())
					if !token.WaitTimeout(3 * time.Second) {
						fmt.Printf("pub qos:2 timeout:%d topic:%s \n", 3, topic)
					}
					atomic.AddUint64(&ops, 1)
				}
			}
		}()

		for i := 0; i < 100000; i++ {
			time.Sleep(time.Second)
			fmt.Println("pub and sub ------>>>", ops)
		}

	}

}

func TestMQTTConnection(t *testing.T) {
	config := Config{
		Host: "emqt1.ahaent.cn",
		Port: 18883,
		//Action:   "pubsub",
		Action:   "pub",
		Topic:    "pubsub/test",
		Username: "",
		Password: "",
		Qos:      0,
		Tls:      false,
		CaCert:   "",
	}
	MQTTConnection(config)
}

func TestMQTTConnectionSub(t *testing.T) {
	config := Config{
		Host: "emqt1.ahaent.cn",
		Port: 18883,
		//Action:   "pubsub",
		Action:   "sub",
		Topic:    "pubsub/test",
		Username: "",
		Password: "",
		Qos:      0,
		Tls:      false,
		CaCert:   "",
	}
	MQTTConnection(config)
}
