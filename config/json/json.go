package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fengfenghuo/go-common-lib/config"
)

func init() {
	config.Register("json", &Config{})
}

// Config ...
type Config struct{}

// Parse returns a ConfigContainer with parsed json config map.
func (conf *Config) Parse(filename string) (config.Configer, error) {
	return loadFromJSON(filename)
}

func loadFromJSON(filename string) (*ConfigEngine, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf ConfigEngine
	err = json.Unmarshal(file, &conf.Data)
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
	val := conf.getData(key)
	if val != nil {
		return config.ParseBool(val)
	}
	return false, fmt.Errorf("not exist key: %q", key)
}

// DefaultBool return the bool value if has no error
// otherwise return the defaultval
func (conf *ConfigEngine) DefaultBool(key string, defaultval bool) bool {
	if v, err := conf.Bool(key); err == nil {
		return v
	}
	return defaultval
}

// Int returns the integer value for a given key.
func (conf *ConfigEngine) Int(key string) (int, error) {
	val := conf.getData(key)
	if val != nil {
		if v, ok := val.(float64); ok {
			return int(v), nil
		}
		return 0, fmt.Errorf("not int value")
	}
	return 0, fmt.Errorf("not exist key:" + key)
}

// DefaultInt returns the integer value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultInt(key string, defaultval int) int {
	if v, err := conf.Int(key); err == nil {
		return v
	}
	return defaultval
}

// Int64 returns the int64 value for a given key.
func (conf *ConfigEngine) Int64(key string) (int64, error) {
	val := conf.getData(key)
	if val != nil {
		if v, ok := val.(float64); ok {
			return int64(v), nil
		}
		return 0, fmt.Errorf("not int64 value")
	}
	return 0, fmt.Errorf("not exist key:" + key)
}

// DefaultInt64 returns the int64 value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultInt64(key string, defaultval int64) int64 {
	if v, err := conf.Int64(key); err == nil {
		return v
	}
	return defaultval
}

// Float returns the float value for a given key.
func (conf *ConfigEngine) Float(key string) (float64, error) {
	val := conf.getData(key)
	if val != nil {
		if v, ok := val.(float64); ok {
			return v, nil
		}
		return 0.0, fmt.Errorf("not float64 value")
	}
	return 0.0, fmt.Errorf("not exist key:" + key)
}

// DefaultFloat returns the float64 value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultFloat(key string, defaultval float64) float64 {
	if v, err := conf.Float(key); err == nil {
		return v
	}
	return defaultval
}

// String returns the string value for a given key.
func (conf *ConfigEngine) String(key string) string {
	val := conf.getData(key)
	if val != nil {
		if v, ok := val.(string); ok {
			return v
		}
	}
	return ""
}

// DefaultString returns the string value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultString(key string, defaultval string) string {
	// TODO FIXME should not use "" to replace non existence
	if v := conf.String(key); v != "" {
		return v
	}
	return defaultval
}

// Strings returns the []string value for a given key.
func (conf *ConfigEngine) Strings(key string) []string {
	stringVal := conf.String(key)
	if stringVal == "" {
		return nil
	}
	return strings.Split(conf.String(key), ";")
}

// DefaultStrings returns the []string value for a given key.
// if err != nil return defaultval
func (conf *ConfigEngine) DefaultStrings(key string, defaultval []string) []string {
	if v := conf.Strings(key); v != nil {
		return v
	}
	return defaultval
}

// SaveConfigFile save the config into file
func (conf *ConfigEngine) SaveConfigFile(filename string) (err error) {
	// Write configuration file by filename.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.MarshalIndent(conf.Data, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// section.key or key
func (conf *ConfigEngine) getData(key string) interface{} {
	if len(key) == 0 {
		return nil
	}

	sectionKeys := strings.Split(key, "::")
	if len(sectionKeys) >= 2 {
		curValue, ok := conf.Data[sectionKeys[0]]
		if !ok {
			return nil
		}
		for _, key := range sectionKeys[1:] {
			if v, ok := curValue.(map[string]interface{}); ok {
				if curValue, ok = v[key]; !ok {
					return nil
				}
			}
		}
		return curValue
	}
	if v, ok := conf.Data[key]; ok {
		return v
	}
	return nil
}
