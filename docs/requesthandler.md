# RequestHandler
RequestHandler类的子类实现Get( ),Post( )等方法处理业务逻辑

## HandlerInterface的结构
```
type HandlerInterface interface {
	Execute([]string)
	Init(HandlerInterface, *HttpRequest, http.ResponseWriter, *Application, Dictionary)
	CallMethod(string, []string)
	Initialize(Dictionary)
	RenderByte(string, map[string]interface{}) []byte
	Prepare()
	Finish()
	GetStatus() int
	HttpError(int, string)
	RaiseHttpError(int, string)
	SetDefaultHeaders()
	FunctionsMap() map[string]interface{}

}
```

## RequestHandler的结构
```
type RequestHandler struct {
	ResponseWriter  http.ResponseWriter
	Request         *HttpRequest
	Status          int
	xsrfToken       string
	application     Application
	contentType     string
	WroteHeader     bool
	delegate HandlerInterface
	RaiseError      bool
	Expires         int
}
```
* RequestHandler实现了HandlerInterface的接口
*  delegate 子类的引用
*  RaiseError 是否主动引发异常，在RaiseHttpError方法中设置为true
* Expires Cookie过期时间

## RequestHandler函数

*  ``GetArgument(name string) string``
	获取请求的参数包括form和url中的参数，如果有多个重名参数取第一个
*  ``GetArguments(name string) []string``
	获取请求的参数包括form和url中的参数, 返回string数组
*  ``GetArgumentDefault(name, _default string) string``
	获取请求的参数包括form和url中的参数，如果有多个重名参数取第一个，如果不存在返回_default的值
*  ``GetQueryArgument(name string) string``
	获取请求的url中的参数，如果有多个重名参数取第一个
*  ``GetQueryArguments(name string) []string ``
	获取请求的url中的参数, 返回string数组
*  ``GetQueryArgumentDefault(name, _default string) string`` 
	获取请求的url中的参数，如果有多个重名参数取第一个，如果不存在返回_default的值
*  ``ReverseUrl(name string, params ...string) string``
	返回命名的handler的url
*  ``GetContentType() string``
	返回请求的内容类型（``application/json``, ``application/xml``, ``text/xml``）
*  ``IsJson() bool``
	是否为json类型
*  ``IsXml() bool``
	是否为xml类型
*  ``IsHtml() bool``
	 是否为html类型
*   ``GetStatus()``
	当前的respose的状态码
*  ``Initialize(params Dictionary)``
	在对象初始化时进行一些先期处理，使用时必须要明确每个值的类型，由于params是``interface{}``的数组，所有必须使用断言。例子如下：
```
type Hello struct {
	...
}
type HelloHandler struct {
	lemon.RequestHandler
	hello Hello
}
lemon.AddRouter("/test", &SubMyhandler{}, map[string]interface{}{"hello": Hello{...}}, "")
func (h *HelloHandler) Initialize(params lemon.Dictionary) {
	h.hello, _ := params["hello"].(Hello)
}
```
*  ``Prepare()``
	主要在Post(),Get()等方法执行前执行，可以做一些预处理
*  ``Finish()``
	在请求结束前执行的函数，可以做一些善后工作
*  ``Clear()``
	这个函数在``Init``中调用，response的header进行初始化
*  ``DeleteHeader(name string)``
	在response的header中删除name的值
*  ``SetHeader(name, value string)``
	在response的header中添加内容
*  ``GetHeader(name string) string``
	获取请求头中的值
*  ``SetDefaultHeaders()``
	在``Clear( )``中调用，设置一些默认的值，可以在子类中覆盖
*  ``Head(args ...string)、 Put(args ...string)、Post(args ...string)、Get(args ...string)、Patch(args ...string)、Options(args ...string、Delete(args ...string)``
	这些方法可以在子类中选择性实现
*  ``GetSecureCookie(key string) string``
	获取安全Cookie, 如果不存在返回""
*  ``SetSecureCookie(name, value string, others map[string]interface{})``
	设置安全Cookie，使用安全Cookie，必须设置``Application.CookieSecret``
	others 的键为："Domain", "expires"(time.time类型), "Max-Age", "Path", "Secure", "HttpOnly"
*  ``ClearCookie(name string)``
	清除Cookie, 只是简单的让Cookie过期
*   ``ClearAllCookie()``
	清除所有Cookie
*  ``WriteHeader(status int)``
	写入response的状态码
*  ``HttpError(status int, message string)``
	  直接向response写入4**, 5**的错误码，以及错误信息
*  ``Write(content []byte)``
	将content信息写入response中
* ``WriteString(str string)``
	向response中写入str信息
*  ``Render(templateName string, context map[string]interface{})``
	渲染模版并且写入response
*  ``RenderByte(templateName string, context map[string]interface{}) []byte``
	渲染模版，如果``Application.IsCustomedTemplate``为true，那么这个函数必须重写
*  ``FunctionsMap() map[string]interface{}``
	可以在这个函数中添加在模版中使用的函数
*  ``Execute(args []string)``
	在这个函数中执行具体的方法 例如：Post方法
*   ``Redirect(url string, status int)``
	重定向函数status必须在300～399之间
*  ``ServeFile(file string)``
	下载文件，调用此函数之前，必须设置好相关的头信息
*  ``XsrfFormHtml() string``
	生成隐藏的form input域
*  ``XsrfToken() string``
	 生成xsrf的token并且添加到安全cookie中
*  ``CheckXsrfCookie() bool``
	检查是否有Xsrftoken

###Xsrf预防
跨站伪造请求(Cross-site request forgery)， 简称为 XSRF，是个性化 Web 应用中常见的一个安全问题。前面的链接也详细讲述了 XSRF 攻击的实现方式。

当前防范 XSRF 的一种通用的方法，是对每一个用户都记录一个无法预知的 cookie 数据，然后要求所有提交的请求中都必须带有这个 cookie 数据。如果此数据不匹配 ，那么这个请求就可能是被伪造的。

Lemon 有内建的 XSRF 的防范机制，要使用此机制，你需要在应用配置中加上 XSRFCookie以及CookieSecret 设定：
```
settings = {
    "CookieSecret": "61oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o/Vo=",
    "login_url": "/login",
    "XSRFCookie": True,
}
lemon.AddRouter("/", &MainHandler{}, lemon.NullDictionary(), "")
```
如果设置了 XSRFCookie，那么 Lemon 的 Web 应用将对所有用户设置一个 _xsrf 的 cookie 值，如果 POST PUT DELET 请求中没有这 个 cookie 值，那么这个请求会被直接拒绝。如果你开启了这个机制，那么在所有 被提交的表单中，你都需要加上一个域来提供这个值。你可以通过在模板中使用 专门的函数 xsrf_form_html() 来做到这一点：
```
<form action="/new_message" method="post">
  {{ xsrf_form_html() }}
  <input type="text" name="message"/>
  <input type="submit" value="Post"/>
</form>
```
如果你提交的是 AJAX 的 POST 请求，你还是需要在每一个请求中通过脚本添加上 _xsrf 这个值。下面是在 FriendFeed 中的 AJAX 的 POST 请求，使用了 jQuery 函数来为所有请求组东添加 _xsrf 值：
```
function getCookie(name) {
    var r = document.cookie.match("\\b" + name + "=([^;]*)\\b");
    return r ? r[1] : undefined;
}

jQuery.postJSON = function(url, args, callback) {
    args._xsrf = getCookie("_xsrf");
    $.ajax({url: url, data: $.param(args), dataType: "text", type: "POST",
        success: function(response) {
        callback(eval("(" + response + ")"));
    }});
};
```
对于 PUT 和 DELETE 请求（以及不使用将 form 内容作为参数的 POST 请求） 来说，你也可以在 HTTP 头中以 X-XSRFToken 这个参数传递 XSRF token。

如果你需要针对每一个请求处理器定制 XSRF 行为，你可以重写 RequestHandler.CheckXsrfCookie()。例如你需要使用一个不支持 cookie 的 API， 你可以通过将 CheckXsrfCookie() 函数设空来禁用 XSRF 保护机制。然而如果 你需要同时支持 cookie 和非 cookie 认证方式，那么只要当前请求是通过 cookie 进行认证的，你就应该对其使用 XSRF 保护机制，这一点至关重要。

# RedirectHandler
这个类主要使用于重定向使用，例子如下：
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
		lemon.AddRouter("/redirect", &lemon.RedirectHandler{}, map[string]interface{}{"Permanent": true, "Url": "http://www.baidu.com"}, ""),
	}

	server := lemon.NewLemon().Instance(handlers, settings)
	server.Listen("", 8080)
	server.Loop()

}
```
如果指定Permanent则为永久重定向，如果Url没有提供，则重定向到"/"。
