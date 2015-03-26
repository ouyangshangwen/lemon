package lemon

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpRequest struct {
	Xheaders       bool
	startTime      time.Time
	finishTime     time.Time
	Request        *http.Request
	QueryArguments url.Values
	FormArguments  url.Values // Parameters from the request body.
	MaxMemory      int64
	Files          map[string][]*multipart.FileHeader // Files uploaded in a multipart form
}

func NewHttpRequest(req *http.Request, xhearders bool, MaxMemory int) *HttpRequest {
	maxMemory := int64(MaxMemory)
	httpRequest := HttpRequest{
		Request:   req,
		Xheaders:  xhearders,
		startTime: time.Now(),
		MaxMemory: maxMemory,
	}
	httpRequest.ParseParams()
	return &httpRequest
}

func (hr *HttpRequest) ParseParams() {
	req := hr.Request
	hr.QueryArguments = req.URL.Query()
	hr.FormArguments = make(url.Values)
	// Parse the body depending on the content type.
	ContentType := ResolveContentType(req)
	switch ContentType {
	case "application/x-www-form-urlencoded":
		// Typical form.
		if err := req.ParseForm(); err != nil {
			lemonLag.Warning(fmt.Sprintf("Error parsing request body:%v", err))
		} else {
			hr.FormArguments = req.Form
		}

	case "multipart/form-data":
		// Multipart form.
		// TODO: Extract the multipart form param so app can set it.
		if err := req.ParseMultipartForm(hr.MaxMemory); err != nil {
			lemonLag.Warning(fmt.Sprintf("Error parsing request body:%v", err))
		} else {
			hr.FormArguments = req.MultipartForm.Value
			hr.Files = req.MultipartForm.File
		}
	default:
		if err := req.ParseForm(); err != nil {
			lemonLag.Warning(fmt.Sprintf("Error parsing request body:%v", err))
		} else {
			hr.FormArguments = req.Form
		}
	}

	for key, values := range hr.QueryArguments {
		hr.FormArguments[key] = append(hr.FormArguments[key], values...)
	}

}

func (hr *HttpRequest) Url() string {
	return hr.Request.URL.Path
}

func (hr *HttpRequest) Protocal() string {
	return hr.Request.Proto // HTTP/1.1 or HTTP/1.0
}

func (hr *HttpRequest) Uri() string {
	return hr.Request.RequestURI
}

func (hr *HttpRequest) Scheme() string {
	var scheme string
	if hr.Request.URL.Scheme != "" {
		scheme = hr.Request.URL.Scheme
	} else if hr.Request.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	if hr.Xheaders == true {
		proto := hr.HeaderDefault("X-Forwarded-Proto", scheme)
		proto = hr.HeaderDefault("X-Scheme", proto)
		if proto == "http" {
			scheme = "http"
		}
		if proto == "https" {
			scheme = proto
		}
	}
	return scheme
}

//contain port
func (hr *HttpRequest) Host() string {
	return hr.Request.Host
}

func (hr *HttpRequest) FullUrl() string {
	return hr.Scheme() + "://" + hr.Host() + hr.Uri()
}

func (hr *HttpRequest) SupportHttp11() bool {
	return "HTTP/1.1" == hr.Protocal()
}

func (hr *HttpRequest) Header(key string) string {
	return hr.Request.Header.Get(key)
}

func (hr *HttpRequest) HeaderDefault(key, defaults string) string {
	value := hr.Request.Header.Get(key)
	if len(value) == 0 {
		return defaults
	} else {
		return value
	}

}

func (hr *HttpRequest) RemoteIP() string {
	var ip string

	raddr := strings.Split(hr.Request.RemoteAddr, ":")
	if len(raddr) > 0 {
		ip = raddr[0]
	} else {
		ip = "127.0.0.1"
	}
	if hr.Xheaders == true {
		ips := hr.Proxy()
		if len(ips) > 0 && ips[0] != "" {
			ip = strings.Split(ips[0], ":")[0]
			ip = hr.HeaderDefault("X-Real-Ip", ip)

		}
	}
	return ip

}

func (hr *HttpRequest) Proxy() []string {
	if ips := hr.Header("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}

func (hr *HttpRequest) Finish() {
	hr.finishTime = time.Now()
}

func (hr *HttpRequest) RequestTime() time.Duration {
	return hr.finishTime.Sub(hr.startTime)
}

func (hr *HttpRequest) IsAjax() bool {
	return hr.Header("X-Requested-With") == "XMLHttpRequest"
}

func (hr *HttpRequest) IsUpload() bool {
	return hr.Header("Content-Type") == "multipart/form-data"
}

func (hr *HttpRequest) Refer() string {
	return hr.Header("Referer")
}

func (hr *HttpRequest) UserAgent() string {
	return hr.Header("User-Agent")
}

//func (hr *HttpRequest) Query(key string) string {
//	return hr.FormArguments.Get(key)
//}

func (hr *HttpRequest) Method() string {
	return hr.Request.Method
}

func (hr *HttpRequest) Body() []byte {
	defer hr.Request.Body.Close()
	requestbody, err := ioutil.ReadAll(hr.Request.Body)
	if err != nil {
		return []byte{}
	} else {
		return requestbody
	}
}

func (hr *HttpRequest) Cookies() []*http.Cookie {
	return hr.Request.Cookies()
}

func (hr *HttpRequest) Cookie(key string) string {
	ck, err := hr.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return ck.Value
}

// Get the content type.
// e.g. From "multipart/form-data; boundary=--" to "multipart/form-data"
// If none is specified, returns "text/html" by default.
func ResolveContentType(req *http.Request) string {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return "text/html"
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}
