package yaml_test

import (
	"fmt"
	"testing"

	"github.com/fengfenghuo/go-common-lib/config"
)

func TestYamlConfig(t *testing.T) {
	conf, err := config.NewConfig("yaml", "conf.yaml")
	if err != nil {
		fmt.Println("NewConfig error: " + err.Error())
		return
	}

	fmt.Println(conf.String("db.module"))
}
