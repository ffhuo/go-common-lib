package json_test

import (
	"fmt"
	"testing"

	"github.com/fengfenghuo/go-common-lib/config"
)

func TestJsonConfig(t *testing.T) {
	conf, err := config.NewConfig("json", "conf.json")
	if err != nil {
		fmt.Println("NewConfig error: " + err.Error())
		return
	}
	fmt.Println(conf.Int("httpport"))
	fmt.Println(conf.String("db::module"))
}
