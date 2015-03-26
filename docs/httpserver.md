# Lemon

Lemon主要继承``http.Server``,Lemon是一个单例对象，对请求事件进行监听，当有请求时，寻找相应的handler进行处理，Lemon可以对以下方式进行监听：http, https, fastcgi(standard I/O, unix, tcp)。

##Lemon 结构

```
type Lemon struct {
	Server  *http.Server
	app     *Application
	address string
	port    int
	network string //used in the mode of fastcgi, value is '(standard I/O)', 'tcp', 'unix'
}
```
*  app 必须实现``ServeHTTP``接口

* address 是监听地址

*  port 是监听端口

*  network 是fastcgi方式必须设置的值，包括 "", "tcp", "unix"

##Lemon函数

* ``func (lem *Lemon) Instance(urlSpecs []UrlSpec, settings map[string]interface{}) *Lemon ``
	这个函数主要的作用时完成初始化工作并返回当前实例，``Application``,在本函数中完成初始化。

*  ``(lem *Lemon) Listen(address string, port int)``
	监听相应的端口与地址，如果，为fastcgi的unix方式，address为unixsocket地址。

*  `` (lem *Lemon) FCGILoop(network string)``
	接受fastcgi方式，参数network的值为：""，标准IO，"tcp"，tcp方式，"unix"，unixsocket方式。

*  ``func (lem *Lemon) Loop()``
	监听http或https方式，如果在settings中有"CertFile"与"KeyFile"则https方式，否则为http方式。
