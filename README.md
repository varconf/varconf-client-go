# varconf-client-go
> 基于go语言的varconf客户端sdk.

![](https://img.shields.io/badge/language-go-cccfff.svg)
[![Build Status](https://travis-ci.org/varconf/varconf-client-go.svg?branch=master)](https://travis-ci.org/varconf/varconf-client-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/varconf/varconf-client-go)](https://goreportcard.com/report/github.com/varconf/varconf-client-go)

## 使用步骤
```
go get github.com/varconf/varconf-client-go
```

## 使用示例
`1.自动配置对象字段`
```go
package main

import (
	"fmt"
	"github.com/varconf/varconf-client-go"
	"log"
	"os"
	"strconv"
)

// 自动注入必须带varconf注解标志
type Test struct {
	Key string `varconf:"key"`
}

// 监听配置变化
func configListener(key, value string, timestamp int64) {
	fmt.Println("key: " + key + " value: " + value + " timestamp: " + strconv.Itoa(int(timestamp)))
}

func main() {
	client, _ := client.NewClient("http://xxx", "your app token")

	test := Test{}

	// add listener
	client.SetListener(configListener)

	// test's filed will change automatic
	client.Watch(&test, 5)
}
```

`2.主动获取对象字段`
```go
package main

import (
	"fmt"
	"github.com/varconf/varconf-client-go"
	"log"
	"os"
	"strconv"
)

func main() {
	client, _ := client.NewClient("http://xxx", "your app token")

    	// 手动拉取配置
	pullKeyResult, _ := client.GetKeyConfig("key", true, 0)
	fmt.Println("1 lastIndex: " + strconv.Itoa(pullKeyResult.RecentIndex))
	fmt.Println("1 key: " + pullKeyResult.Data.Key + " value: " + pullKeyResult.Data.Value + " timestamp: " + strconv.Itoa(int(pullKeyResult.Data.Timestamp)))

	pullKeyResult, _ = client.GetKeyConfig("key", true, pullKeyResult.RecentIndex)
	fmt.Println("2 lastIndex: " + strconv.Itoa(pullKeyResult.RecentIndex))
	fmt.Println("2 key: " + pullKeyResult.Data.Key + " value: " + pullKeyResult.Data.Value + " timestamp: " + strconv.Itoa(int(pullKeyResult.Data.Timestamp)))
}
```

