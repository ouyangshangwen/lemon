package lemon

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/ouyangshangwen/lemon/utils"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const (
	applicationJson = "application/json"
	applicationXml  = "application/xml"
	textXml         = "text/xml"
)

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
	CheckXsrfCookie() bool
}

type RequestHandler struct {
	ResponseWriter http.ResponseWriter
	Request        *HttpRequest
	Status         int
	XsrfToken      string
	application    Application
	contentType    string
	WroteHeader    bool
	delegate       HandlerInterface
	RaiseError     bool
	Expires        int
}

func (rh *RequestHandler) Init(self HandlerInterface, request *HttpRequest,
	rw http.ResponseWriter, app *Application, params Dictionary) {
	rh.ResponseWriter = rw
	rh.Request = request
	rh.application = *app
	rh.Status = 200
	rh.delegate = self
	rh.Clear()
	rh.WroteHeader = false
	rh.delegate.Initialize(params)
}

func (rh *RequestHandler) GetDelegate() HandlerInterface {
	return rh.delegate
}

func (rh *RequestHandler) GetApplicaion() Application {
	return rh.application
}

//Returns the value of the argument with the given name.
//
//If the argument is not present, return an empty string
func (rh *RequestHandler) GetArgument(name string) string {
	return rh._getArgument(name, "")
}

//Returns the value of the argument with the given name.
//
//default must be provided
func (rh *RequestHandler) GetArgumentDefault(name, _default string) string {
	return rh._getArgument(name, _default)
}

//Returns the value of the argument with the given name.
//
//If the argument is not present, returns an empty list
func (rh *RequestHandler) GetArguments(name string) []string {
	return rh._getArguments(name)
}

func (rh *RequestHandler) _getArgument(name, _default string) string {
	value := rh._getArguments(name)
	if len(value) == 0 {
		return _default
	}
	return value[0]
}

func (rh *RequestHandler) _getArguments(name string) []string {
	values, ok := rh.Request.FormArguments[name]
	if !ok {
		return []string{}
	} else {
		return values
	}
}

//Returns the value of the argument with the given name,
//from the request query string
//If default is not provided, default is ""
func (rh *RequestHandler) GetQueryArgument(name string) string {
	return rh._getQueryArgument(name, "")
}

//Returns the value of the argument with the given name,
// from the request query string
//default must be provided
func (rh *RequestHandler) GetQueryArgumentDefault(name, _default string) string {
	return rh._getQueryArgument(name, _default)
}

//Returns a list of the query arguments with the given name.
//
//If the argument is not present, returns an empty list
func (rh *RequestHandler) GetQueryArguments(name string) []string {
	return rh._getQueryArguments(name)
}

func (rh *RequestHandler) _getQueryArgument(name, _default string) string {
	value := rh._getQueryArguments(name)
	if len(value) == 0 {
		return _default
	}
	return value[0]
}

func (rh *RequestHandler) _getQueryArguments(name string) []string {
	values, ok := rh.Request.QueryArguments[name]
	if !ok {
		return []string{}
	} else {
		return values
	}
}

//Alias for `Application.ReverseUrl`
func (rh *RequestHandler) ReverseUrl(name string, params ...string) string {
	return rh.application.ReverseUrl(name, params...)
}

func (rh *RequestHandler) GetContentType() string {
	contentType := rh.GetHeader("Content-Type")
	if contentType == "" {
		return "text/html"
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}
func (rh *RequestHandler) IsJson() bool {
	return rh.GetContentType() == applicationJson
}

func (rh *RequestHandler) IsXml() bool {
	return rh.GetContentType() == applicationXml
}

func (rh *RequestHandler) IsHtml() bool {
	return rh.GetContentType() == textXml
}
func (rh *RequestHandler) GetStatus() int {
	return rh.Status
}

//Hook for subclass initialization.
//
//A dictionary passed as the third argument of a url spec will be
//supplied as keyword arguments to Initialize()
func (rh *RequestHandler) Initialize(params Dictionary) {

}

//Called at the beginning of a request before  `get`/`post`/etc.
//
//  Override this method to perform common initialization regardless
//  of the request method.
func (rh *RequestHandler) Prepare() {

}

func (rh *RequestHandler) Finish() {

}

//Resets all headers and content for this response
func (rh *RequestHandler) Clear() {
	rh.SetHeader("Server", rh.application.ServerName)
	rh.SetHeader("Content-Type", "text/html; charset=UTF-8")
	gmttime := utils.AppendTime([]byte{}, time.Now())
	rh.SetHeader("Date", string(gmttime))
	rh.delegate.SetDefaultHeaders()
	if !rh.Request.SupportHttp11() {
		connHeader := rh.GetHeader("Connection")
		if len(connHeader) > 0 && strings.ToLower(connHeader) == "keep-alive" {
			rh.SetHeader("Connection", "Keep-Alive")
		}

	}
	rh.Status = 200

}

//delete an outgoing header.
func (rh *RequestHandler) DeleteHeader(name string) {
	rh.ResponseWriter.Header().Del(name)
}

//Sets the given response header name and value.
func (rh *RequestHandler) SetHeader(name, value string) {
	rh.ResponseWriter.Header().Set(name, value)
}

func (rh *RequestHandler) GetHeader(name string) string {
	return rh.Request.Header(name)
}

func (rh *RequestHandler) SetDefaultHeaders() {

}

//func (rh *RequestHandler) Head(args ...string) {
//	http.Error(rh.ResponseWriter, "Method Not Allowed", 405)
//}
//
//func (rh *RequestHandler) Put(args ...string) {
//	//http.Error(rh.ResponseWriter, "Method Not Allowed", 405)
//	rh.Status = 405
//	panic("Method Not Allowed")
//}
//
//func (rh *RequestHandler) Post(args ...string) {
//	//http.Error(rh.ResponseWriter, "Method Not Allowed", 405)
//	rh.Status = 405
//	panic("Method Not Allowed")
//}
//
//func (rh *RequestHandler) Get(args ...string) {
//	//http.Error(rh.ResponseWriter, "Method Not Allowed", 405)
//	rh.Status = 405
//	panic("Method Not Allowed")
//}
//
//func (rh *RequestHandler) Patch(args ...string) {
//	//http.Error(rh.ResponseWriter, "Method Not Allowed", 405)
//	rh.Status = 405
//	panic("Method Not Allowed")
//}
//
//func (rh *RequestHandler) Options(args ...string) {
//	//http.Error(rh.ResponseWriter, "Method Not Allowed", 405)
//	rh.Status = 405
//	panic("Method Not Allowed")
//}
//
//func (rh *RequestHandler) Delete(args ...string) {
//	//http.Error(rh.ResponseWriter, "Method Not Allowed", 405)
//	rh.Status = 405
//	panic("Method Not Allowed")
//}

// GetSecureCookie returns decoded cookie value from encoded browser cookie values.
func (rh *RequestHandler) GetSecureCookie(key string) string {
	secureCookie, ok := rh.getSecureCookie(key)
	if ok {
		return secureCookie
	} else {
		return ""
	}
}

// Set Secure cookie for response.
func (rh *RequestHandler) SetSecureCookie(name, value string, others map[string]interface{}) {
	secret := rh.application.CookieSecret
	if len(secret) == 0 {
		secret = "cookive secret"
	}
	cookie := rh.createdSignedValue(secret, name, value)
	rh.SetCookie(name, cookie, others)
}

// Get cookie from request by a given key.
// It's alias of HttpRequest.Cookie.
func (rh *RequestHandler) GetCookie(key string) string {
	return rh.Request.Cookie(key)
}

// Set cookie for response.
func (rh *RequestHandler) SetCookie(name string, value string, others map[string]interface{}) {
	rh.setCookie(name, value, others)
}

//Deletes the cookie with the given name.
func (rh *RequestHandler) ClearCookie(name string) {
	rh.setCookie(name, "", map[string]interface{}{"expires": time.Now().UTC().Add(time.Duration(-365*24*60*60) * time.Second)})
}

//Deletes all the cookies the user sent with this request.
func (rh *RequestHandler) ClearAllCookie() {
	for _, cookie := range rh.Request.Cookies() {
		rh.ClearCookie(cookie.Name)
	}
}

var COOKIENAME = []string{"Domain", "expires", "Max-Age", "Path", "Secure", "HttpOnly"}

func (rh *RequestHandler) setCookie(name string, value string, others map[string]interface{}) {
	var b bytes.Buffer
	var cookieValue string
	fmt.Fprintf(&b, "%s=%s", sanitizeName(name), sanitizeValue(value))
	_, ok := others["Path"]
	if !ok {
		others["Path"] = "/"
	}
	for _, name := range COOKIENAME {

		if value, ok := others[name]; ok {

			switch reflect.ValueOf(value).Type().Kind() {
			case reflect.Int:
				intValue, _ := value.(int)
				cookieValue = strconv.Itoa(intValue)
			case reflect.Struct:
				timeslice := []byte{}
				expires, ok := value.(time.Time)
				if !ok {
					lemonLag.Error("cookie's expires must be the type of time.Time")
					panic("cookie's expires must be the type of time.Time")
				}
				timeslice = utils.AppendTime([]byte{}, expires)
				cookieValue = string(timeslice)
			case reflect.String:
				cookieValue, _ = value.(string)

			}

			if name == "Secure" || name == "HttpOnly" {
				fmt.Fprintf(&b, "; %s", name)
				continue
			}
			cookieValue = sanitizeValue(cookieValue)
			fmt.Fprintf(&b, "; %s=%s", name, cookieValue)
		}

	}

	rh.ResponseWriter.Header().Add("Set-Cookie", b.String())
}

// Get secure cookie from request by a given key.
func (rh *RequestHandler) getSecureCookie(key string) (string, bool) {
	secret := rh.application.CookieSecret
	if len(secret) == 0 {
		secret = "cookive secret"
	}
	val := rh.GetCookie(key)
	if val == "" {
		return "", false
	}

	parts := strings.SplitN(val, "|", 3)

	if len(parts) != 3 {
		return "", false
	}

	vs := parts[0]
	timestamp := parts[1]
	sig := parts[2]

	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return "", false
	}
	res, _ := base64.URLEncoding.DecodeString(vs)
	return string(res), true
}

var cookieNameSanitizer = strings.NewReplacer("\n", "-", "\r", "-")

func sanitizeName(n string) string {
	return cookieNameSanitizer.Replace(n)
}

var cookieValueSanitizer = strings.NewReplacer("\n", " ", "\r", " ", ";", " ")

func sanitizeValue(v string) string {
	return cookieValueSanitizer.Replace(v)
}

func (rh *RequestHandler) createdSignedValue(secret, name, value string) string {
	vs := base64.URLEncoding.EncodeToString([]byte(value))
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	return cookie
}

func (rh *RequestHandler) decodeSignedValue(secret, value string) string {
	if value == "" {
		return ""
	}

	parts := strings.SplitN(value, "|", 3)

	if len(parts) != 3 {
		return ""
	}

	vs := parts[0]
	timestamp := parts[1]
	sig := parts[2]

	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return ""
	}
	res, _ := base64.URLEncoding.DecodeString(vs)
	return string(res)
}

func (rh *RequestHandler) WriteHeader(status int) {
	rh.WroteHeader = true
	rh.ResponseWriter.WriteHeader(status)
}

func (rh *RequestHandler) HttpError(status int, message string) {
	rh.WriteHeader(status)
	rh.WriteStringOnly(message)

}

func (rh *RequestHandler) WriteStringOnly(message string) {
	rh.ResponseWriter.Write([]byte(message))

}

func (rh *RequestHandler) Write(content []byte) {

	if !rh.WroteHeader {
		rh.WriteHeader(rh.Status)
	}
	if rh.application.IsGzip == true && rh.Request.Header("Accept-Encoding") != "" {
		output_writer := rh.ResponseWriter.(io.Writer)
		splitted := strings.SplitN(rh.Request.Header("Accept-Encoding"), ",", -1)
		encodings := make([]string, len(splitted))

		for i, val := range splitted {
			encodings[i] = strings.TrimSpace(val)
		}
		for _, val := range encodings {
			if val == "gzip" {
				rh.SetHeader("Content-Encoding", "gzip")
				output_writer, _ = gzip.NewWriterLevel(rh.ResponseWriter, gzip.BestSpeed)
				defer output_writer.(*gzip.Writer).Close()
				break
			} else if val == "deflate" {
				rh.SetHeader("Content-Encoding", "deflate")
				output_writer, _ = flate.NewWriter(rh.ResponseWriter, flate.BestSpeed)
				defer output_writer.(*flate.Writer).Close()
				break
			}

			_, err := output_writer.Write(content)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	} else {

		rh.SetHeader("Content-Length", strconv.Itoa(len(content)))
		rh.ResponseWriter.Write(content)
		return
	}

}

func (rh *RequestHandler) WriteString(str string) {
	rh.WriteHeader(rh.Status)
	rh.ResponseWriter.Write([]byte(str))
}

func (rh *RequestHandler) Render(templateName string, context map[string]interface{}) {
	html := rh.delegate.RenderByte(templateName, context)
	rh.Write(html)
	//fmt.Println(html)

}

func (rh *RequestHandler) RenderByte(templateName string, context map[string]interface{}) []byte {

	if rh.application.IsCustomedTemplate {
		rh.Status = 500
		panic("if `Application.IsCustomedTemplate is true`, must overwrite the method of RenderByte ")
	}
	namespace := rh.GetTemplateNamespace()
	for key, value := range namespace {
		context[key] = value
	}

	template := rh.CreateTemplateLoader(templateName)
	newbytes := bytes.NewBufferString("")

	err := template.ExecuteTemplate(newbytes, templateName, context)
	if err != nil {
		lemonLag.Info(err)
	}
	content, err := ioutil.ReadAll(newbytes)
	if err != nil {
	}
	return content

}

func (rh *RequestHandler) CreateTemplateLoader(templateName string) *template.Template {
	template, ok := Templates.Templates[templateName]
	if !ok {
		rh.Status = 404
		panic("no this template: " + templateName)

	}
	funcMaps := AddFuncMap(rh.delegate.FunctionsMap())
	template = template.Funcs(funcMaps)
	return template
}

func (rh *RequestHandler) GetTemplatePath() string {
	return rh.application.TemplatePath

}
func (rh *RequestHandler) FunctionsMap() map[string]interface{} {
	return make(map[string]interface{}, 0)
}

func (rh *RequestHandler) GetTemplateNamespace() map[string]interface{} {
	namespace := map[string]interface{}{
		"Handler":      rh.delegate,
		"Request":      rh.Request,
		"XsrfFormHtml": template.HTML(rh.XsrfFormHtml()),
	}
	return namespace
}

func (rh *RequestHandler) recoverFromPanic() {
	if err := recover(); err != nil {
		if rh.RaiseError {
			return
		}

		if rh.application.Debug {
			//panic(err)
            debugStack := debug.Stack()
            resposeErr := fmt.Sprintf(`<html><title>Error</title><body>
                <pre style=\"word-wrap: break-word; white-space: pre-wrap;\" >%s</pre>
                </body></html>`, string(debugStack))
            rh.HttpError(rh.Status, resposeErr)
            fmt.Println(string(debugStack))
			return
		}
		status := rh.Status
		if status > 499 && status < 600 {
			rh.HttpError(status, "server error")
			fmt.Println(err)
			return
		}
		if status == 404 {
			rh.HttpError(404, "not found the page")
			return
		}
	}
}

var SUPPORTEDMETHOD = []string{"GET", "HEAD", "POST", "DELETE", "PATCH", "PUT", "OPTIONS"}
var XSRFMETHOD = []string{"GET", "HEAD", "OPTIONS"}

func (rh *RequestHandler) Execute(args []string) {
	defer rh.recoverFromPanic()
	if rh.checkNotMethod(SUPPORTEDMETHOD) {
		rh.RaiseHttpError(405, "Method not allow")

	}
	if rh.checkNotMethod(XSRFMETHOD) && rh.application.XSRFCookie {
		rh.delegate.CheckXsrfCookie()
	}
	rh.delegate.Prepare()
	method := rh.Request.Method()
	method = strings.ToLower(method)
	method = strings.Title(method)

	rh.CallMethod(method, args)

	rh.Request.Finish()
    logInfo := fmt.Sprintf(" %d %s %s %s", rh.Status, rh.Request.Method(), rh.Request.Url(), rh.Request.RequestTime())

	lemonLag.Info(logInfo)
	rh.delegate.Finish()

}

func (rh *RequestHandler) checkNotMethod(methods []string) bool {
	for _, method := range methods {
		if strings.EqualFold(rh.Request.Method(), method) {
			return false
		}
	}
	return true
}

func (rh *RequestHandler) CallMethod(methodName string, args []string) {
	var ptr reflect.Value
	var value reflect.Value
	var finalMethod reflect.Value
	value = reflect.ValueOf(rh.delegate)
	// if we start with a pointer, we need to get value pointed to
	// if we start with a value, we need to get a pointer to that value
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(rh.delegate))
		temp := ptr.Elem()
		temp.Set(value)
	}

	// check for method on value
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	if finalMethod.IsValid() {
		length := finalMethod.Type().NumIn()
		in := make([]reflect.Value, length)
		if length == len(args) {
			for i, arg := range args {
				in[i] = reflect.ValueOf(arg)
			}
		} else {
			lemonLag.Error("args not match")
			panic("args not match")
		}
		finalMethod.Call(in)
		return
	}else {
        rh.HttpError(405, "method not allowed") 
	    lemonLag.Error("no this method")
        return
    }
}

func (rh *RequestHandler) Redirect(url string, status int) {
	if status < 300 || status > 399 {
		lemonLag.Error("status must 300~399")
		panic("status must 300~399")
	}

	rh.Status = status
	rh.SetHeader("Location", url)
	rh.delegate.Finish()
	rh.WriteHeader(status)
}

func (rh *RequestHandler) ServeFile(file string) {
	http.ServeFile(rh.ResponseWriter, rh.Request.Request, file)

}

func (rh *RequestHandler) RaiseHttpError(status int, message string) {
	rh.WriteHeader(status)
	rh.WriteStringOnly(message)
	rh.RaiseError = true
	panic("raise error by self")
}

// XsrfFormHtml writes an input field contains xsrf token value.
func (rh *RequestHandler) XsrfFormHtml() string {
	rh.GetXsrfToken()

	return `<input type="hidden" name="_xsrf" value="` +
		rh.XsrfToken + `">`

}

// GetXsrfToken creates a xsrf token string and returns.
// you can set global expire in Application or set in your handler default 60 second
func (rh *RequestHandler) GetXsrfToken() string {
	intExpire := rh.application.Expires
	var timeduration int
	if rh.Expires != 0 {
		timeduration = rh.Expires
	} else if intExpire == 0 {
		timeduration = 60

	} else {
		timeduration = intExpire
	}
	if rh.XsrfToken == "" {
		token := rh.GetSecureCookie("_xsrf")
		if len(token) == 0 {
			token = string(utils.RandomCreateBytes(32))
			expires := time.Now().UTC().Add(time.Duration(timeduration) * time.Second)
			rh.SetSecureCookie("_xsrf", token, map[string]interface{}{"expires": expires})
		}
		rh.XsrfToken = token
	}
	return rh.XsrfToken
}

// CheckXsrfCookie checks xsrf token in this request is valid or not.
// the token can provided in request header "X-Xsrftoken" and "X-CsrfToken"
// or in form field value named as "_xsrf".
func (rh *RequestHandler) CheckXsrfCookie() bool {
	token := rh.GetArgumentDefault("_xsrf", "")
	if token == "" {
		token = rh.Request.Header("X-Xsrftoken")
	}
	if token == "" {
		token = rh.Request.Header("X-Csrftoken")
	}
	if token == "" {
		rh.RaiseHttpError(403, "'_xsrf' argument missing from POST")
	} else if rh.GetXsrfToken() != token {
		rh.RaiseHttpError(403, "XSRF cookie does not match POST argument")
	}
	return true
}

type RedirectHandler struct {
	RequestHandler
	Url string
	//Permanent bool
}

func (redirecthandler *RedirectHandler) Initialize(params Dictionary) {
	_, ok := params["Permanent"]
	if ok {
		redirecthandler.Status = 301
	} else {
		redirecthandler.Status = 302
	}

	Url, ok := params["Url"]
	if ok {
		redirecthandler.Url = Url.(string)
	} else {
		redirecthandler.Url = "/"
	}

}

func (redirecthandler *RedirectHandler) Get() {

	redirecthandler.Redirect(redirecthandler.Url, redirecthandler.Status)
}
