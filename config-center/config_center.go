package center

import (
	// "bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/go-sfox-lib/sfox/rpc"
)

// Topic the topic struct
type Topic struct {
	Name       string
	Version    int
	Content    json.RawMessage
	Monitors   []*func(string, json.RawMessage) int
	Subscribed bool
}

type topicsContains map[string]*Topic

type CenterClient struct {
	url              string
	mqttURL          string
	topics           topicsContains
	subscriptions    topicsContains
	mqttClientID     string
	mqttClient       MQTT.Client
	reconnectedTimer *time.Ticker
	messageArrived   MQTT.MessageHandler
}

// NewInstance is the init function
func NewInstance(serverURL, mqttURL string) *CenterClient {
	client := CenterClient{
		url:           serverURL,
		mqttURL:       mqttURL,
		topics:        topicsContains{},
		subscriptions: topicsContains{},
		mqttClientID:  uniqueID(),
	}

	client.messageArrived = func(mqttClient MQTT.Client, msg MQTT.Message) {
		log.Printf("接收消息tipic: %s", msg.Topic())
		log.Printf("接收消息content: %s", msg.Payload())

		type MessageData struct {
			Content json.RawMessage `json:"content"`
			Version int             `json:"version"`
		}
		var message MessageData
		err := json.Unmarshal(msg.Payload(), &message)
		if err != nil {
			log.Println("messageArrived Unmarshal msg error: " + err.Error())
			return
		}

		content := message.Content
		// mqtt数据为string类型，去开始和结束的‘“’
		if message.Content[0] == '"' {
			content = message.Content[1 : len(message.Content)-1]
		}

		topic, ok := client.subscriptions[msg.Topic()]
		if ok {
			client.updateTopicContent(topic, content, message.Version)
		}
	}

	client.startSyncTimer(time.Minute)

	return &client
}

func (client *CenterClient) SubscribeAndQuery(topicName string, monitor func(string, json.RawMessage) int) json.RawMessage {
	topic, ok := client.topics[topicName]
	if !ok {
		var err error
		topic, err = client.createTopic(topicName)
		if err != nil {
			log.Printf("center:SubscribeAndQuery:createTopic error: " + err.Error())
			return nil
		}
	}

	if !topic.Subscribed {
		err := client.subscribeTopic(topic)
		if err != nil {
			client.checkStartReconnectTimer(30 * time.Second)
		}
	}

	topic.Monitors = append(topic.Monitors, &monitor)

	return topic.Content
}

func (client *CenterClient) Unsubscribe(topicName string, monitor func(string, json.RawMessage) int) error {
	topic, ok := client.topics[topicName]
	if ok {
		if err := client.doUnsubscribe(topic, monitor); err != nil {
			return err
		}
	}
	return nil
}

func (client *CenterClient) updateTopicContent(topic *Topic, content json.RawMessage, version int) {
	if version > topic.Version {
		log.Printf("startSyncTimer: response: %s", string(content[:]))

		isSuccess := true
		for _, monitor := range topic.Monitors {
			rv := (*monitor)(topic.Name, content)
			if rv != 0 {
				isSuccess = false
			}
		}

		if isSuccess {
			topic.Content = content
			topic.Version = version
		}
	}
}

func (client *CenterClient) createTopic(topicName string) (*Topic, error) {
	url := client.url + "/config_server/topics/" + strings.Split(topicName, ".")[1]

	res, err := rpc.SendHttpRequest(url)
	if err != nil {
		return nil, fmt.Errorf("SendHttpRequest error: %s", err.Error())
	}
	log.Println("query topic data: ", string(res[:]))
	var topic Topic
	err = json.Unmarshal(res, &topic)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal topic data error: " + err.Error())
	}
	topic.Name = topicName
	client.subscriptions[topicName] = &topic
	return &topic, nil
}

func (client *CenterClient) subscribeTopic(topic *Topic) error {
	if client.mqttClient == nil {
		log.Println("mqtt-connect: " + client.mqttURL + " clientID: " + client.mqttClientID)
		client.buildConnectMqttClient()
	}

	if !client.mqttClient.IsConnected() {
		if err := client.doConnectMqtt(); err != nil {
			return err
		}
	}

	if err := client.doSubscribeMqttTopic(topic.Name); err != nil {
		return fmt.Errorf("doSubscribeMqttTopic error: " + err.Error())
	}

	topic.Subscribed = true
	return nil
}

func (client *CenterClient) doUnsubscribe(topic *Topic, monitor func(string, json.RawMessage) int) error {
	for index, temp := range topic.Monitors {
		if &monitor == temp {
			if len(topic.Monitors) < index {
				topic.Monitors = append(topic.Monitors[:index], topic.Monitors[index+1:]...)
			} else {
				topic.Monitors = topic.Monitors[:index]
			}
		}
	}

	if len(topic.Monitors) == 0 {
		if err := client.doUnsubscribeMqttTopic(topic.Name); err != nil {
			return fmt.Errorf("doUnsubscribeMqttTopic error: " + err.Error())
		}

		delete(client.topics, topic.Name)
		log.Printf("取消订阅: " + topic.Name)
	}
	return nil
}

func (client *CenterClient) startSyncTimer(delay time.Duration) {
	time.AfterFunc(delay, func() {
		defer client.startSyncTimer(delay)

		if len(client.subscriptions) == 0 {
			return
		}

		// 判断是否正常连接配置服务和MQTT
		if !client.mqttClient.IsConnected() {
			for _, topic := range client.subscriptions {
				topic.Subscribed = false
			}

			err := client.startReConnect()
			if err != nil {
				log.Println(err)
				return
			}
		}

		var topicArray = []string{}
		for _, topic := range client.subscriptions {
			isExist := false
			topicName := strings.Split(topic.Name, ".")[1]
			for _, data := range topicArray {
				if topicName == data {
					isExist = true
					break
				}
			}

			if !isExist {
				topicArray = append(topicArray, topicName)
			}
		}

		req, err := json.Marshal(topicArray)
		if err != nil {
			log.Printf("startSyncTimer: Marshal error: %s", err.Error())
			return
		}

		urlValues := url.Values{}
		urlValues.Add("topics", string(req[:]))

		// log.Printf("urlEncode: %s", urlValues.Encode())

		url := client.url + "/config_server/versions?" + urlValues.Encode()

		res, err := rpc.SendHttpRequest(url)
		if err != nil {
			log.Printf("startSyncTimer: SendHttpRequest error: %s", err.Error())
			return
		}

		type MessageData struct {
			Topic   string          `json:"topic"`
			Content json.RawMessage `json:"content"`
			Version int             `json:"version"`
		}

		var topics []MessageData
		err = json.Unmarshal(res, &topics)
		if err != nil {
			log.Printf("startSyncTimer: Unmarshal %s, error: %s", string(res[:]), err.Error())
			return
		}

		for _, data := range topics {
			topic, ok := client.subscriptions[data.Topic]
			if ok {
				client.updateTopicContent(topic, data.Content, data.Version)
			}
		}
	})
}

func (client *CenterClient) checkStartReconnectTimer(delay time.Duration) error {
	if client.reconnectedTimer == nil {
		client.reconnectedTimer = time.NewTicker(delay)

		go func() {
			defer func() {
				if client.reconnectedTimer != nil {
					client.reconnectedTimer.Stop()
					client.reconnectedTimer = nil
				}
			}()

			for {
				select {
				case _ = <-client.reconnectedTimer.C:
					if err := client.startReConnect(); err != nil {
						log.Println(err)
						continue
					}
					return
				}
			}
		}()
	}
	return nil
}

func (client *CenterClient) startReConnect() error {
	if !client.mqttClient.IsConnected() {
		if err := client.doConnectMqtt(); err != nil {
			return fmt.Errorf("checkStartReconnectTimer: doConnectMqtt error: %s", err.Error())
		}
	}

	if err := client.checkAndSubscribeAll(); err != nil {
		return fmt.Errorf("checkStartReconnectTimer: checkAndSubscribeAll error: %s", err.Error())
	}

	return nil
}

func (client *CenterClient) checkAndSubscribeAll() error {
	for _, topic := range client.subscriptions {
		if !topic.Subscribed {
			if err := client.doSubscribeMqttTopic(topic.Name); err != nil {
				log.Printf("checkAndSubscribeAll: doSubscribeMqttTopic : %s error: %s", topic.Name, err.Error())
				continue
			}
			topic.Subscribed = true
		}
	}
	return nil
}

func (client *CenterClient) buildConnectMqttClient() error {
	opt := MQTT.NewClientOptions().AddBroker(client.mqttURL).SetClientID(client.mqttClientID)
	opt.SetDefaultPublishHandler(client.messageArrived)
	client.mqttClient = MQTT.NewClient(opt)
	return nil
}

func (client *CenterClient) doSubscribeMqttTopic(topicName string) error {
	if client.mqttClient == nil {
		return fmt.Errorf("mqtt is not connected")
	}

	if token := client.mqttClient.Subscribe(topicName, 0, nil); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (client *CenterClient) doUnsubscribeMqttTopic(topicName string) error {
	if client.mqttClient == nil {
		return fmt.Errorf("mqtt is not connected")
	}

	if token := client.mqttClient.Unsubscribe(topicName); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (client *CenterClient) doConnectMqtt() error {
	if token := client.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func uniqueID() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}

	h := md5.New()
	h.Write([]byte(b))
	return hex.EncodeToString(h.Sum(nil))
}
