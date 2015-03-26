#  HttpRequest

HttpRequest是request的请求对象，在这个类中，主要是对http的请求header与body进行处理。

## HttpRequest结构

```
type HttpRequest struct {
	Xheaders bool
	startTime      time.Time
	finishTime     time.Time
	Request        *http.Request
	QueryArguments url.Values
	FormArguments  url.Values // Parameters from the request body.
	MaxMemory      int64
	Files          map[string][]*multipart.FileHeader // Files uploaded in a multipart form
}
```
*  Xheaders 用于代理或负载均衡，只有Xheaders为true时，从``X-Real-Ip``/``X-Forwarded-For``取RemoteIP ，从``X-Scheme``/``X-Forwarded-Proto`` 取 scheme
*  startTime 请求开始时间
*  finishTime 请求结束时间
* Request http.Request指针
* QueryArguments url中的查询参数
* FormArguments body体中取出的参数，包括QueryArguments中的参数
* MaxMemory 上传文件最大值
* Files 上传文件对象数组
* ``func NewHttpRequest(req *http.Request, xhearders bool, MaxMemory int) *HttpRequest``,初始化HttpRequest

## HttpRequest 函数分析

* ``(hr *HttpRequest) ParseParams()``
	根据请求的方法以及form表单，初始化QueryArguments，FormArguments，Files。

*  ``func (hr *HttpRequest) Protocal() string``
	返回http版本（HTTP/1.1 或 HTTP/1.0）
*  ``func (hr *HttpRequest) Url() string``
	返回request的``net/http http.Request.URL.Path``
*  ``func (hr *HttpRequest) Protocal() string``
		返回request的``net/http http.Request.Proto``
* ``func (hr *HttpRequest) Scheme() string``
		返回request的``net/http http.Request.URL.Scheme`` ，如果Xheaders为true会根据请求头中的``X-Forwarded-Proto``或``X-Scheme``返回Scheme
* ``func (hr *HttpRequest) Host() string``
		返回request的``net/http http.Request.Host``
* ``func (hr *HttpRequest) FullUrl() string``
	返回完整的url
*  ``func (hr *HttpRequest) SupportHttp11() bool``
	如果request版本为http1.1返回true，否则返回false
*  ``hr *HttpRequest) Header(key string) string``
	取header中的值
*  ``func (hr *HttpRequest) HeaderDefault(key, defaults string) string``
	带有默认值的函数
*  ``(hr *HttpRequest) RemoteIP() string``
		返回请求IP，如果Xheaders为true，尝试从代理的header中取包括``X-Real-Ip X-Forwarded-For``

*  ``func (hr *HttpRequest) Proxy() []string``
	返回header中``X-Forwarded-For``的数组形式
*  ``func (hr *HttpRequest) Finish()``
	请求结束时间
*  ``(hr *HttpRequest) RequestTime() time.Duration``
	返回整个请求的时间
*  ``func (hr *HttpRequest) Refer() string``
	 返回上一个链接地址
*  ``func (hr *HttpRequest) UserAgent() string``
	返回用户代理头
*  ``func (hr *HttpRequest) Method() string``
	返回请求方法
*  ``func (hr *HttpRequest) Body() []byte``
	返回body体的字节数组
*  ``func (hr *HttpRequest) Cookies() []*http.Cookie``
	返回请求Cookies
*  ``func (hr *HttpRequest) Cookie(key string) string``
	获取Cookie

