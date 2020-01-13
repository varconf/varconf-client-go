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
	pullKeyResult, _ := client.GetKeyConfig("key", true, 0)
	fmt.Println("1 lastIndex: " + strconv.Itoa(pullKeyResult.RecentIndex))
	fmt.Println("1 key: " + pullKeyResult.Data.Key + " value: " + pullKeyResult.Data.Value + " timestamp: " + strconv.Itoa(int(pullKeyResult.Data.Timestamp)))

	pullKeyResult, _ = client.GetKeyConfig("key", true, pullKeyResult.RecentIndex)
	fmt.Println("2 lastIndex: " + strconv.Itoa(pullKeyResult.RecentIndex))
	fmt.Println("2 key: " + pullKeyResult.Data.Key + " value: " + pullKeyResult.Data.Value + " timestamp: " + strconv.Itoa(int(pullKeyResult.Data.Timestamp)))

	// test's filed will change automatic
	client.Watch(&test, 5)
}
