package rpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func SendHttpRequest(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, fmt.Errorf("response error: error code : %d", res.StatusCode)
}

func SendHttpRequestWithData(msg []byte, url string) error {
	reqNew := bytes.NewBuffer(msg)
	request, err := http.NewRequest("POST", url, reqNew)
	if err != nil {
		return err
	}

	request.Header.Set("Content-type", "application/json")

	client := &http.Client{Timeout: time.Duration(3 * time.Second)}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(body[:]))
	} else {
		fmt.Println(response)
	}
	return nil
}
