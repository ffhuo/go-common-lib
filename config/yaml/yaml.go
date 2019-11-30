package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fengfenghuo/go-common-lib/config"
	"gopkg.in/yaml.v2"
)

func init() {
	config.Register("yaml", &Config{})
}

// Config is a yaml config parser and implements Config interface.
type Config struct{}

// Parse returns a ConfigContainer with parsed yaml config map.
func (yaml *Config) Parse(filename string) (y config.Configer, err error) {
	y, err = loadFromYaml(filename)
	return
}

func loadFromYaml(filename string) (*ConfigEngine, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf ConfigEngine
	err = yaml.Unmarshal(file, &conf.Data)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// ConfigEngine ...
type ConfigEngine struct {
	Data map[string]interface{}
}

// Bool returns the boolean value for a given key.
func (conf *ConfigEngine) Bool(key string) (bool, error) {
	v, err := conf.getData(key)
	if err != nil {
		return false, err
	}
	return config.ParseBool(v)
}

// DefaultBool return the bool value if has no error
// otherwise return the defaultval
func (conf *ConfigEngine) DefaultBool(key string, defaultval bool) bool {
	v, err := conf.Bool(key)
	if err != nil {
		return defaultval
	}
	return v
}

// Int returns the integer value for a given key.
func (conf *ConfigEngine) Int(key string) (int, error) {
	if v, err := conf.getData(key); err != nil {
		return 0, err
	} else if vv, ok := v.(int); ok {
		return vv, nil
	} else if vv, ok := v.(int64); ok {
		return int(vv), nil
	}
	return 0, fmt.Errorf("not int value")
}

// DefaultInt returns the integer value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultInt(key string, defaultval int) int {
	v, err := conf.Int(key)
	if err != nil {
		return defaultval
	}
	return v
}

// Int64 returns the int64 value for a given key.
func (conf *ConfigEngine) Int64(key string) (int64, error) {
	if v, err := conf.getData(key); err != nil {
		return 0, err
	} else if vv, ok := v.(int64); ok {
		return vv, nil
	}
	return 0, fmt.Errorf("not bool value")
}

// DefaultInt64 returns the int64 value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultInt64(key string, defaultval int64) int64 {
	v, err := conf.Int64(key)
	if err != nil {
		return defaultval
	}
	return v
}

// Float returns the float value for a given key.
func (conf *ConfigEngine) Float(key string) (float64, error) {
	if v, err := conf.getData(key); err != nil {
		return 0.0, err
	} else if vv, ok := v.(float64); ok {
		return vv, nil
	} else if vv, ok := v.(int); ok {
		return float64(vv), nil
	} else if vv, ok := v.(int64); ok {
		return float64(vv), nil
	}
	return 0.0, fmt.Errorf("not float64 value")
}

// DefaultFloat returns the float64 value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultFloat(key string, defaultval float64) float64 {
	v, err := conf.Float(key)
	if err != nil {
		return defaultval
	}
	return v
}

// String returns the string value for a given key.
func (conf *ConfigEngine) String(key string) string {
	if v, err := conf.getData(key); err == nil {
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	return ""
}

// DefaultString returns the string value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultString(key string, defaultval string) string {
	v := conf.String(key)
	if v == "" {
		return defaultval
	}
	return v
}

// Strings returns the []string value for a given key.
func (conf *ConfigEngine) Strings(key string) []string {
	v := conf.String(key)
	if v == "" {
		return nil
	}
	return strings.Split(v, ";")
}

// DefaultStrings returns the []string value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultStrings(key string, defaultval []string) []string {
	v := conf.Strings(key)
	if v == nil {
		return defaultval
	}
	return v
}

// SaveConfigFile save the config into file
func (conf *ConfigEngine) SaveConfigFile(filename string) (err error) {
	// Write configuration file by filename.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := yaml.Marshal(conf.Data)
	if err != nil {
		return err
	}
	_, err = f.Write(r)
	return err
}

func (conf *ConfigEngine) getData(key string) (interface{}, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	keys := strings.Split(key, ".")
	tmpData := conf.Data

	for idx, k := range keys {
		if v, ok := tmpData[k]; ok {
			switch v.(type) {
			case map[string]interface{}:
				{
					tmpData = v.(map[string]interface{})
					if idx == len(keys)-1 {
						return tmpData, nil
					}
				}
			case map[interface{}]interface{}:
				{
					tmpData = make(map[string]interface{})
					for k, value := range v.(map[interface{}]interface{}) {
						if keyTemp, ok := k.(string); ok {
							tmpData[keyTemp] = value
						}
					}
				}
			default:
				{
					return v, nil
				}

			}
		}
	}
	return nil, fmt.Errorf("not exist key %q", key)
}
