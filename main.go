package main

import (
	"fmt"
	"time"
	"varconf-client-go/client"
)

type Test struct {
	Key string `varconf:"1"`
}

func main() {
	client := client.Client{
		Url:   "http://127.0.0.1:8088",
		Token: "50:a68a0e61-2380-435d-9a4b-87fb93777536",
	}

	var test Test
	go func() {
		for {
			fmt.Println("key:" + test.Key)
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()

	client.Watch(&test)
}
