package center_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fengfenghuo/go-common-lib/config-center"
)

func callBack(topicName string, content json.RawMessage) {
	fmt.Println(topicName, content)
}
func TestConfigCenter(t *testing.T) {
	client := center.NewInstance("http://127.0.0.1:20080", "tcp://221.228.197.195:1883")
	data := client.SubscribeAndQuery("cfg.MaintainMail", callBack)
	fmt.Printf("topic: cfg/notic, content: %v", data)

	var stopChan = make(chan string)
	for {
		select {
		case msg := <-stopChan:
			fmt.Println("stop" + msg)
			return
		}
	}
}
