# varconf-client-go
> 基于go语言的varconf客户端sdk.

![](https://img.shields.io/badge/language-go-cccfff.svg)
[![Build Status](https://travis-ci.org/varconf/varconf-client-go.svg?branch=master)](https://travis-ci.org/varconf/varconf-client-go)

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

// must config with varconf's tag
type Test struct {
	Key string `varconf:"key"`
}

// listener is valid with watch mode
func configListener(key, value string, timestamp int64) {
	fmt.Println("key: " + key + " value: " + value + " timestamp: " + strconv.Itoa(int(timestamp)))
}

func main() {
	client, _ := client.NewClient("http://xxx", "your app token", log.New(os.Stdout, "Info: ", log.Ltime|log.Lshortfile))

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
	client, _ := client.NewClient("http://xxx", "your app token", log.New(os.Stdout, "Info: ", log.Ltime|log.Lshortfile))

	// pull key
	configValue, lastIndex, _ := client.GetKeyConfig("key", true, 0)
	configValue, lastIndex, _ = client.GetKeyConfig("key", true, lastIndex)
	fmt.Println("lastIndex: " + strconv.Itoa(lastIndex))
	fmt.Println("key: " + configValue.Key + " value: " + configValue.Value + " timestamp: " + strconv.Itoa(int(configValue.Timestamp)))
}
```

