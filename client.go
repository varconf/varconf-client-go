package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
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
	url      string
	token    string
	listener Listener
	logger   *log.Logger
}

type Listener func(string, string, int64)

func NewClient(url, token string, logger *log.Logger) (*Client, error) {
	client := &Client{url: url, token: token, logger: logger}
	return client, nil
}

func (_self *Client) Watch(obj interface{}, errSleep int) {
	lastIndex := 0
	for {
		// poll config
		configMap, recentIndex, err := _self.GetAppConfig(true, lastIndex)
		if err != nil {
			if _self.logger != nil {
				_self.logger.Println("varconf client poll config error! detail: " + err.Error())
			}
			time.Sleep(time.Duration(errSleep) * time.Second)
			continue
		}

		// reflect data
		_, err = _self.reflect(obj, configMap)
		if err != nil {
			if _self.logger != nil {
				_self.logger.Println("varconf client reflect data error! detail: " + err.Error())
			}
			time.Sleep(time.Duration(errSleep) * time.Second)
			continue
		}

		// refresh index
		lastIndex = recentIndex
	}
}

func (_self *Client) SetListener(listener Listener) {
	_self.listener = listener
}

func (_self *Client) GetAppConfig(longPull bool, lastIndex int) (map[string]*ConfigValue, int, error) {
	url := _self.url + "/api/config" + "?token=" + _self.token
	if longPull {
		url = url + "&longPull=true&lastIndex=" + strconv.Itoa(lastIndex)
	}

	result, code, err := _self.get(url)
	if code != http.StatusOK || err != nil {
		return nil, -1, errors.New("request error, status: " + strconv.Itoa(code))
	}

	var appResult PullAppResult
	if err := json.Unmarshal([]byte(result), &appResult); err != nil {
		return nil, -1, errors.New("decode error, detail: " + err.Error())
	}
	return appResult.Data, appResult.RecentIndex, nil
}

func (_self *Client) GetKeyConfig(key string, longPull bool, lastIndex int) (*ConfigValue, int, error) {
	url := _self.url + "/api/config/" + key + "?token=" + _self.token
	if longPull {
		url = url + "&longPull=true&lastIndex=" + strconv.Itoa(lastIndex)
	}

	result, code, err := _self.get(url)
	if code != http.StatusOK || err != nil {
		return nil, 0, err
	}

	var keyResult PullKeyResult
	if err := json.Unmarshal([]byte(result), &keyResult); err != nil {
		return nil, 0, err
	}
	return keyResult.Data, keyResult.RecentIndex, nil
}

func (_self *Client) reflect(obj interface{}, configMap map[string]*ConfigValue) (bool, error) {
	// check param
	if obj == nil || configMap == nil {
		return false, errors.New("param can't null")
	}

	// reflect data
	rVal := reflect.ValueOf(obj)
	rType := reflect.TypeOf(obj)
	if rType.Kind() == reflect.Ptr {
		rVal = rVal.Elem()
		rType = rType.Elem()
	} else {
		return false, errors.New("obj must be ptr to struct")
	}

	// fill with data
	for i := 0; i < rType.NumField(); i++ {
		rT := rType.Field(i)
		rV := rVal.Field(i)

		// get field config
		tag := rT.Tag.Get(VARCONF)
		if tag == "" {
			continue
		}
		configValue, ok := configMap[tag]
		if !ok || configValue == nil {
			continue
		}

		// set filed data
		data := configValue.Value
		if rT.Type.Kind() == reflect.Bool {
			val, _ := strconv.ParseBool(data)
			rV.SetBool(val)
		} else if rT.Type.Kind() == reflect.String {
			rV.SetString(data)
		} else if rT.Type.Kind() == reflect.Int32 || rT.Type.Kind() == reflect.Int64 {
			val, _ := strconv.ParseInt(data, 10, 64)
			rV.SetInt(val)
		} else if rT.Type.Kind() == reflect.Float32 || rT.Type.Kind() == reflect.Float64 {
			val, _ := strconv.ParseFloat(data, 64)
			rV.SetFloat(val)
		} else {
			return false, errors.New("not support " + rT.Type.Kind().String())
		}

		// notify listener
		if _self.listener != nil {
			_self.listener(configValue.Key, configValue.Value, configValue.Timestamp)
		}
	}

	return true, nil
}

func (_self *Client) get(url string) (string, int, error) {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return "", -1, err
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
			return "", -1, err
		}
	}
	return result.String(), code, nil
}
