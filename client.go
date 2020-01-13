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

// VARCONF is annotation tag
const (
	VARCONF = "varconf"
)

// ConfigValue configuration details
type ConfigValue struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// PullKeyResult is api result
type PullKeyResult struct {
	Data        *ConfigValue `json:"data"`
	RecentIndex int          `json:"recentIndex"`
}

// PullAppResult is api result
type PullAppResult struct {
	Data        map[string]*ConfigValue `json:"data"`
	RecentIndex int                     `json:"recentIndex"`
}

// Client must init with varconf's url and app's access token
type Client struct {
	url      string
	token    string
	listener Listener
	logger   *log.Logger
}

// Listener is callback function for configuration changes
type Listener func(string, string, int64)

// NewClient is client init method
func NewClient(url, token string) (*Client, error) {
	client := &Client{url: url, token: token}
	return client, nil
}

// SetLogger set the logger for client
func (_self *Client) SetLogger(logger *log.Logger)  {
	_self.logger = logger
}

// SetListener set the listener for watch configuration
func (_self *Client) SetListener(listener Listener) {
	_self.listener = listener
}

// Watch add a object to watch,
// Client will automatically change the field's value
func (_self *Client) Watch(obj interface{}, errSleep int) {
	lastIndex := 0
	for {
		// poll config
		pullAppResult, err := _self.GetAppConfig(true, lastIndex)
		if err != nil {
			if _self.logger != nil {
				_self.logger.Println("varconf client poll config error! detail: " + err.Error())
			}
			time.Sleep(time.Duration(errSleep) * time.Second)
			continue
		}

		// reflect data
		_, err = _self.reflect(obj, pullAppResult.Data)
		if err != nil {
			if _self.logger != nil {
				_self.logger.Println("varconf client reflect data error! detail: " + err.Error())
			}
			time.Sleep(time.Duration(errSleep) * time.Second)
			continue
		}

		// refresh index
		lastIndex = pullAppResult.RecentIndex
	}
}

// GetAppConfig manually pull all configuration
func (_self *Client) GetAppConfig(longPull bool, lastIndex int) (*PullAppResult, error) {
	url := _self.url + "/api/config" + "?token=" + _self.token
	if longPull {
		url = url + "&longPull=true&lastIndex=" + strconv.Itoa(lastIndex)
	}

	result, code, err := _self.get(url)
	if code != http.StatusOK || err != nil {
		return nil, errors.New("request error, status: " + strconv.Itoa(code))
	}

	var appResult PullAppResult
	if err := json.Unmarshal([]byte(result), &appResult); err != nil {
		return nil, errors.New("decode error, detail: " + err.Error())
	}
	return &appResult, nil
}

// GetKeyConfig manually pull the configuration of the specified key
func (_self *Client) GetKeyConfig(key string, longPull bool, lastIndex int) (*PullKeyResult, error) {
	url := _self.url + "/api/config/" + key + "?token=" + _self.token
	if longPull {
		url = url + "&longPull=true&lastIndex=" + strconv.Itoa(lastIndex)
	}

	result, code, err := _self.get(url)
	if code != http.StatusOK || err != nil {
		return nil, err
	}

	var keyResult PullKeyResult
	if err := json.Unmarshal([]byte(result), &keyResult); err != nil {
		return nil, err
	}
	return &keyResult, nil
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
