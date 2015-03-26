#  Application

Application 实现了 ``ServeHTTP`` 接口，Application的主要目的：

-  是通过request的请求，找到相应的``RequestHandler`` 的子类进行处理
-  做一些全局的设置
##Application结构
```
type Application struct {
	ExtraParams        map[string]interface{}
	XSRFCookie         bool   // flag of enable xsrf default false
	Debug              bool   // shorthand for serveral debug default true
	CookieSecret       string // used by RequestHandler.GetSecureCookie and RequestHandler.SetCecureCookie to sign cookies
	Expires            int    // the secret cookie expiration time after Expires seconds
	TemplatePath       string // Directory containing template files
	LeftBraces         string // the left mark of template veriable default "{{"
	RightBraces        string // the right mark of template veriable default "}}"
	Handlers           []HostPattern
	DefaultHost        string
	StaticPath         string //Directory from which static files will be served
	IsGzip             bool   // if or not use gzip compress in response
	NameHandlers       map[string]UrlSpec
	AbsWorkPath        string        // the absolute path of current workspace
	MaxMemory          int           //defalut 64MB
	ReadTimeOut        time.Duration // maximum duration before timing out read of the request
	WriteTimeOut       time.Duration // maximum duration before timing out write of the response
	CertFile           string
	KeyFile            string
	IsCustomedTemplate bool   // if true use customed template engine, must implement the method RenderByte of RequestHandler
	ServerName         string // server name exported in response header.
	Xheaders           bool
	NUMCPU             int // the count of used cup default 1
}
```
## Application中提供的默认参数
-  XSRFCookie ``bool`` 类型
	是否使用安全Cookie，默认 false
-  Debug ``bool`` 类型
	是否为开发模式，默认为true
-  CookieSecret ``string`` 类型
	安全Cookie的签名，配合XSRFCookie使用， 当XSRFCookie为true，CookieSecret不能为空
-  Expires ``int`` 类型
	Cookie过期时间，默认0，单位：秒
-  TemplatePath ``string`` 类型        
	模版路径，默认当前工作目录的 ``template/``
-  LeftBraces, RightBraces ``string`` 类型
	模版变量标识，默认``{{``, ``}}``
-  DefaultHost ``string`` 类型
	 默认的Host
-  StaticPath ``string`` 类型
	静态文件路径，默认当前工作目录的 ``static/``
-  IsGzip ``bool`` 类型
	是否对response进行压缩，默认false
-  MaxMemory ``int`` 类型
	上传文件最多值，默认64M
-  ReadTimeOut ``time.Duration`` 类型
	request请求过期时间，默认0
-  WriteTimeOut ``time.Duration`` 类型
	response响应过期时间，默认0
-  CertFile ``string`` 类型
	数字证书地址， 默认""
-  KeyFile ``string`` 类型
	数字签名地址，默认"", 当KeyFile或CertFile默认使用https请求方式
-  IsCustomedTemplate ``bool`` 类型
	是否自定义模版引擎，默认false, 如果IsCustomedTemplate 为true时，全局变量Templates(``*Template``类型)，此时如果需要渲染html模版，需要重写``RequestHandler.RenderByte`` 
-  ServerName ``string`` 类型
	http响应头中Server的名称，默认LemonServer
-  NUMCPU ``int`` 类型
	启动cpu核心数，默认1

## Application 的函数

 -  ``Init(urlSpecs []UrlSpec, settings map[string]interface{})`` 主要进行默认配置，生成Requesthandler 列表。
urlSpecs: url 到requesthandler的映射列表
settings: 修改默认参数的值
 -  ``AddHandlers(hostPattern string, hosthandlers []UrlSpec)`` 
	 将hostPattern 映射到 hosthandlers
 - ``ServeHTTP(rw http.ResponseWriter, r *http.Request)``
	 实现``net/http`` 的 ``ServeHTTP``接口，指定具体的处理``RequestHandler``
 - ``ReverseUrl(name string, params ...string) string``
 - ``parseSettings(settings map[string]interface{})``
	内部函数，处理``Init``函数中的settings，如果settings中的关键字是``Application``的属性，则转化为响应类型，如果关键字不在``Application``的属性中，settings中的值保存在``Application.ExtraParams``。在 ``RequestHandler``的方法中，如下使用
	``attributename := application.ExtraParams[key].(type)``获取，settings的key的值。

#HostPattern

指定host到UrlSpec列表的映射

##HostPattern结构

```
type HostPattern struct {
	hostCompiled *regexp.Regexp
	handlers     []UrlSpec
}
```

- ``func NewHostPattern(hostPattern string, hosthandlers []UrlSpec) HostPattern``
	创建HostPattern对象

#UrlSpec

指定url到requesthandler的映射

##UrlSpec结构

```
type UrlSpec struct {
	Name         string
	pattern      string
	HandlerClass HandlerInterface
	Kwargs       Dictionary
	Regexps      *regexp.Regexp
	path         string
	GroupCount   int
}
```

- ``func (urls *UrlSpec) Reverse(args ...string) (string, error)``
	通过名称获取requesthandler的url
- ``func NewUrlSpec(pattern string, handlerclass HandlerInterface, name string, params Dictionary) UrlSpec`` 
	创建UrlSpec

##快捷函数与结构

- type Dictionary map[string]interface{}
	只是为``map[string]interface{}``取别名
- ``func NullDictionary() Dictionary ``
	返回Null 的 Dictionary

- ``func AddRouter(pattern string, handler HandlerInterface, params Dictionary, name string) UrlSpec``
	NewUrlSpec的别名函数




	
	

   