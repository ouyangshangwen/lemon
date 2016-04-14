Lemon Web Server
=======

lemon 是go语言开发的web框架，是对``net/http``的简单封装，它的设计完全来自于python 的web框架tornado，同时在beego，revel中吸收了一些思想。
lemon的设计目标，设计最小，最精简的框架，所有无关的部分，采用组件的形式使用。

简单的web应用
-------

这是一个简单的 “Hello World“ web 应用

```
 package main

import (
	"lemon"
)

func main() {

	settings := map[string]interface{}{
		"IsGzip": false,
		"NUMCPU":     8,
	}
	handlers := []lemon.UrlSpec{
		lemon.AddRouter("/", &HelloWorldHandler{}, lemon.NullDictionary(), ""),
	}

	server := lemon.NewLemon().Instance(handlers, settings)
	server.Listen("", 8080)
	server.Loop()

}

type HelloWorldHandler struct {
	lemon.RequestHandler
}

func (hw *HelloWorldHandler) Get() {
	hw.WriteString("Hello, world")
}
```
