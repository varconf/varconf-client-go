package main

import (
	"fmt"
	"github.com/varconf/varconf-client-go"
	"log"
	"os"
	"strconv"
)

// must config with varconf's tag
type Test struct {
	Key string `varconf:"key"`
}

func configListener(key, value string, timestamp int64) {
	fmt.Println("key: " + key + " value: " + value + " timestamp: " + strconv.Itoa(int(timestamp)))
}

func main() {
	client, _ := client.NewClient("varconf's url", "your app token", log.New(os.Stdout, "Info: ", log.Ltime|log.Lshortfile))

	test := Test{}

	// add listener
	client.SetListener(configListener)

	// pull key
	configValue, lastIndex, _ := client.GetKeyConfig("key", true, 0)
	configValue, lastIndex, _ = client.GetKeyConfig("key", true, lastIndex)
	fmt.Println("lastIndex: " + strconv.Itoa(lastIndex))
	fmt.Println("key: " + configValue.Key + " value: " + configValue.Value + " timestamp: " + strconv.Itoa(int(configValue.Timestamp)))

	// test's filed will change automatic
	client.Watch(&test, 5)
}
