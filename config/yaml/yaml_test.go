package yaml_test

import (
	"fmt"
	"testing"

	"github.com/go-sfox-lib/sfox/config"
)

func TestYamlConfig(t *testing.T) {
	conf, err := config.NewConfig("yaml", "conf.yaml")
	if err != nil {
		fmt.Println("NewConfig error: " + err.Error())
		return
	}

	fmt.Println(conf.String("db.module"))
}
