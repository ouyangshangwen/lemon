# StaticFileHandler
这个类是用于静态文件的请求，在生产环境最后，使用代理来代替，这样能提高效率。

## StaticFileHandler 结构
```
type StaticFileHandler struct {
	RequestHandler
	ModifiedTime time.Time
	CACHEMAXAGE  int
	root         string
}
```
*  root静态文件的根目录
* CACHEMAXAGE 最大缓存时间
* ModifiedTime 文件修改时间

## 示例
```
package main

import (
	"fmt"
	"lemon"
)

func main() {

	settings := map[string]interface{}{
		"IsGzip":       false,
		"CookieSecret": "bfalfbalfblafbalfeeee",
		"XSRFCookie":   true,
		"NUMCPU":       8,
	}
	handlers := []lemon.UrlSpec{
		lemon.AddRouter("/test/static/img/(.*)", &lemon.StaticFileHandler{}, map[string]interface{}{"path": "/var/www"}, ""),
	}

	server := lemon.NewLemon().Instance(handlers, settings)
	server.Listen("", 8080)
	server.Loop()
}
```