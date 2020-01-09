package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

const (
	VARCONF = "varconf"
)

type ConfigValue struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

type PullKeyResult struct {
	Data        *ConfigValue `json:"data"`
	RecentIndex int          `json:"recentIndex"`
}

type PullAppResult struct {
	Data        map[string]*ConfigValue `json:"data"`
	RecentIndex int                     `json:"recentIndex"`
}

type Client struct {
	Url      string
	Token    string
	watching bool
}

func (_self *Client) Watch(obj interface{}) {
	_self.watching = true
	lastIndex := 0
	for {
		configMap, recentIndex := _self.GetAppConfig(true, lastIndex)
		if configMap == nil {
			continue
		}

		_self.reflect(obj, configMap)
		lastIndex = recentIndex

		if !_self.watching {
			break
		}
	}
}

func (_self *Client) Stop() {
	_self.watching = false
}

func (_self *Client) GetAppConfig(longPull bool, lastIndex int) (map[string]*ConfigValue, int) {
	url := _self.Url + "/api/config" + "?token=" + _self.Token
	if longPull {
		url = url + "&longPull=true&lastIndex=" + strconv.Itoa(lastIndex)
	}

	result, code := _self.get(url)
	if code != http.StatusOK {
		return nil, 0
	}

	var appResult PullAppResult
	if err := json.Unmarshal([]byte(result), &appResult); err != nil {
		return nil, 0
	}
	return appResult.Data, appResult.RecentIndex
}

func (_self *Client) GetKeyConfig(key string, longPull bool, lastIndex int) (*ConfigValue, int) {
	url := _self.Url + "/api/config/" + key + "?token=" + _self.Token
	if longPull {
		url = url + "&longPull=true&lastIndex=" + strconv.Itoa(lastIndex)
	}

	result, code := _self.get(url)
	if code != http.StatusOK {
		return nil, 0
	}

	var keyResult PullKeyResult
	if err := json.Unmarshal([]byte(result), &keyResult); err != nil {
		return nil, 0
	}
	return keyResult.Data, keyResult.RecentIndex
}

func (_self *Client) reflect(obj interface{}, configMap map[string]*ConfigValue) (bool, error) {
	rVal := reflect.ValueOf(obj)
	rType := reflect.TypeOf(obj)
	if rType.Kind() == reflect.Ptr {
		rVal = rVal.Elem()
		rType = rType.Elem()
	} else {
		return false, errors.New("obj must be ptr to struct")
	}

	for i := 0; i < rType.NumField(); i++ {
		tag := rType.Field(i).Tag.Get(VARCONF)
		if tag == "" {
			continue
		}
		configValue, ok := configMap[tag]
		if ok {
			if configValue != nil {
				rVal.Field(i).Set(reflect.ValueOf(configValue.Value))
			}
		}
	}
	return true, nil
}

func (_self *Client) get(url string) (string, int) {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	result := bytes.NewBuffer(nil)
	var buffer [512]byte
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	return result.String(), code
}
