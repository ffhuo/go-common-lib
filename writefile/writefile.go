package writefile

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func WriteFileToJSON(pathFile string, data interface{}) error {
	_, err := os.Stat(pathFile)
	if err != nil {
		if os.IsExist(err) {
			os.Remove(pathFile)
		}
	}
	f, err := os.Create(pathFile)
	if err != nil {
		return err
	}

	defer f.Close()

	byt, _ := json.Marshal(data)
	f.Write(byt)
	return nil
}

func WriteFileToCSV(pathFile string, data interface{}, example interface{}) error {
	_, err := os.Stat(pathFile)
	if err != nil {
		if os.IsExist(err) {
			os.Remove(pathFile)
		}
	}
	f, err := os.Create(pathFile)
	if err != nil {
		return err
	}

	defer f.Close()

	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)

	title := []string{}

	rt := reflect.TypeOf(example).Elem()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		tempValue := field.Tag.Get("json")
		if tempValue == "" {
			continue
		}
		title = append(title, tempValue)
	}

	w.Write(title)

	dataValues := reflect.ValueOf(data)
	if dataValues.Kind() != reflect.Slice && dataValues.Kind() != reflect.Array {
		return fmt.Errorf("data is not a slice or array")
	}

	for i := 0; i < dataValues.Len(); i++ {
		value := dataValues.Index(i)
		var record []string
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			var changeValue string
			switch field.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
				changeValue = strconv.FormatInt(field.Int(), 10)
			case reflect.String:
				changeValue = field.String()
			default:
				fmt.Println("write file type unKnow: ", field.Kind())
			}

			record = append(record, changeValue)
		}
		w.Write(record)
	}
	w.Flush()
	return nil
}
