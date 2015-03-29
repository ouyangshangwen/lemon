package lemon

import (
	"errors"
	"fmt"
	"github.com/ouyangshangwen/lemon/utils"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var Templates *Template

// Application generally consists of one or more RequestHandler subclasses,
// an Application object which routes incoming requests to handlers.
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

func NewApplication() *Application {
	return &Application{}

}

func (app *Application) Init(urlSpecs []UrlSpec, settings map[string]interface{}) {
	app.setDefaultValue()
	app.parseSettings(settings)
	numCPU := runtime.NumCPU()
	if app.NUMCPU != 1 {
		if int(app.NUMCPU) > numCPU {
			runtime.GOMAXPROCS(numCPU)
		} else {
			if app.NUMCPU < 1 {
				app.NUMCPU = 1
			}
			runtime.GOMAXPROCS(int(app.NUMCPU))
		}
	}

	workPath, _ := os.Getwd()
	AbsWorkPath, _ := filepath.Abs(workPath)
	app.AbsWorkPath = AbsWorkPath
	app.NameHandlers = map[string]UrlSpec{}

	if len(app.StaticPath) == 0 {
		app.StaticPath = filepath.Join(AbsWorkPath, "/static")
		statics := [3]string{"/static/(.*)", "/(favicon.ico)", "/(robots.txt)"}
		for _, static := range statics {
			staticHander := &StaticFileHandler{}
			//			var self HandlerInterface
			//			self = staticHander
			urlSpec := AddRouter(static,
				staticHander, map[string]interface{}{"path": app.StaticPath}, "")
			//Handlers = append(Handlers, urlHandler)
			urlSpecs = append(urlSpecs, urlSpec)
		}

	}

	if app.Handlers == nil {
		app.AddHandlers(".*$", urlSpecs)
	}
	if !app.IsCustomedTemplate {
		Templates = TemplateInit(app.LeftBraces, app.RightBraces)
		var templatePath string
		if len(app.TemplatePath) == 0 {
			templatePath = filepath.Join(app.AbsWorkPath, "/templates/")
		} else {
			templatePath = app.TemplatePath
		}
		Templates.BuildTemplate(templatePath)
	}

}

func (app *Application) setDefaultValue() {

	workPath, _ := os.Getwd()
	AbsWorkPath, _ := filepath.Abs(workPath)
	app.AbsWorkPath = AbsWorkPath
	app.XSRFCookie = false
	app.Debug = true
	app.IsCustomedTemplate = false

	app.MaxMemory = 1 << 26 //64MB
	app.IsGzip = false
	app.Expires = 0

	app.LeftBraces = "{{"
	app.RightBraces = "}}"
	app.ServerName = "LemonServer"
	app.NUMCPU = 1
	app.ReadTimeOut = time.Duration(0) * time.Second
	app.WriteTimeOut = time.Duration(0) * time.Second

}

func (app *Application) parseSettings(settings map[string]interface{}) {
	app.ExtraParams = make(map[string]interface{})
	for key, value := range settings {
		appType := reflect.TypeOf(app).Elem()
		appValue := reflect.ValueOf(app).Elem()
		_, ok := appType.FieldByName(key)
		if ok {
			if key == "ReadTimeOut" {
				readTimeOut, _ := value.(int)
				app.ReadTimeOut = time.Duration(readTimeOut) * time.Second
				continue
			}
			if key == "WriteTimeOut" {
				writeTimeOut, _ := value.(int)
				app.WriteTimeOut = time.Duration(writeTimeOut) * time.Second
				continue
			}
			switch appValue.FieldByName(key).Kind() {
			case reflect.Int:
				IntValue, _ := value.(int)
				appValue.FieldByName(key).SetInt(int64(IntValue))
			case reflect.String:
				StringValue, _ := value.(string)
				appValue.FieldByName(key).SetString(StringValue)
			case reflect.Bool:
				BoolVaue, _ := value.(bool)
				appValue.FieldByName(key).SetBool(BoolVaue)
			}
		} else {
			app.ExtraParams[key] = value
		}

	}
}

//Appends the given handlers to our handler list.
//
//Host patterns are processed sequentially in the order they were
//added. All matching patterns will be considered.
func (app *Application) AddHandlers(hostPattern string, hosthandlers []UrlSpec) {
	if !strings.HasSuffix(hostPattern, "$") {
		hostPattern = hostPattern + "$"
	}
	//	hostPatternEntry := HostPattern{}
	//	hostComiled, _ := regexp.Compile(hostPattern)
	//	//fmt.Println(error)
	//	hostPatternEntry.hostCompiled = hostComiled
	//	hostPatternEntry.handlers = hosthandlers
	hostPatternEntry := NewHostPattern(hostPattern, hosthandlers)
	for _, spec := range hosthandlers {
		name := spec.Name
		if len(name) != 0 {
			_, ok := app.NameHandlers[name]
			if ok {
				errLog := fmt.Sprintf("Multiple handlers named %s; replacing previous value", name)
				lemonLag.Error(errLog)
				panic(errLog)
			}
			app.NameHandlers[name] = spec
		}

	}

	length := len(app.Handlers)
	insertIndex := length - 1
	if length > 0 && app.Handlers[insertIndex].Pattern() == ".*$" {
		app.Handlers = append(app.Handlers[:length], append([]HostPattern{hostPatternEntry}, app.Handlers[length:]...)...)
	} else {
		app.Handlers = append(app.Handlers, hostPatternEntry)
	}

}

func (app *Application) _getHostHandler(request *HttpRequest) []UrlSpec {
	hostList := strings.Split(request.Host(), ":")
	host := hostList[0]
	var matches []UrlSpec
	for _, hostpattern := range app.Handlers {
		match := hostpattern.hostCompiled.MatchString(host)
		if match {
			matches = hostpattern.handlers
		}

	}
	//Look for default host if not behind load balancer (for debugging)
	if len(matches) == 0 && len(request.HeaderDefault("X-Real-Ip", "")) == 0 {
		for _, hostpattern := range app.Handlers {
			match := hostpattern.hostCompiled.MatchString(app.DefaultHost)
			if match {
				matches = hostpattern.handlers
			}
		}
	}
	return matches
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (app *Application) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	//time.Sleep(100 * time.Microsecond)
	var handler HandlerInterface

	request := NewHttpRequest(r, app.Xheaders, app.MaxMemory)
	handlers := app._getHostHandler(request)
	var args []string

	kwargs := NullDictionary()
	if len(handlers) == 0 {

		redirecthandler := &RedirectHandler{}

		kwargs["Url"] = "http://" + app.DefaultHost + r.RequestURI
		kwargs["permanent"] = true
		redirecthandler.Init(redirecthandler, request, rw, app, kwargs)
		//redirecthandler.Url = "http://" + app.DefaultHost + "/"
		handler = redirecthandler

	} else {
		for _, spec := range handlers {
			match := spec.Regexps.MatchString(request.Url())
			if match {
				handler = spec.HandlerClass
				args = spec.Regexps.FindStringSubmatch(request.Url())[1:]
				kwargs = spec.Kwargs
				break
			}
		}

	}

	if handler == nil {
		app.NotFound(rw, r)
	} else {
		var instanceHandlerInterface HandlerInterface
		instanceValue := reflect.ValueOf(handler)
		instanceType := reflect.Indirect(instanceValue).Type()
		instance := reflect.New(instanceType)
		instanceHandlerInterface, ok := instance.Interface().(HandlerInterface)

		if !ok {
			panic("is not HandlerInterface")
		}

		instanceHandlerInterface.Init(instanceHandlerInterface, request, rw, app, kwargs) // #
		instanceHandlerInterface.Execute(args)
		handler = nil
	}

}

func (app *Application) NotFound(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Server", app.ServerName)
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	gmttime := utils.AppendTime([]byte{}, time.Now())
	rw.Header().Set("Date", string(gmttime))
	if r.Proto != "HTTP/1.1" {
		connHeader := r.Header.Get("Connection")
		if len(connHeader) > 0 && strings.ToLower(connHeader) == "keep-alive" {
			rw.Header().Set("Connection", "Keep-Alive")
		}

	}
	http.NotFound(rw, r)
}

//	Returns a URL path for handler named ``name``
//
//	The handler must be added to the application as a named `URLSpec`.
//
//	Args will be substituted for capturing groups in the `URLSpec` regex.
func (app *Application) ReverseUrl(name string, params ...string) string {
	urlSpec, ok := app.NameHandlers[name]
	if !ok {
		errLog := fmt.Sprintf("%s not found in named urls", name)
		lemonLag.Error(errLog)
		panic(errLog)
	}
	url, err := urlSpec.Reverse(params...)
	if err != nil {
		errLog := fmt.Sprintf("%v", err.Error())
		panic(errLog)
	}
	return url
}

//Specifies mappings between hosts and UrlSpecs.
type HostPattern struct {
	hostCompiled *regexp.Regexp
	handlers     []UrlSpec
}

func (hp *HostPattern) Pattern() string {
	return hp.hostCompiled.String()
}
func NewHostPattern(hostPattern string, hosthandlers []UrlSpec) HostPattern {
	hostPatternEntry := HostPattern{}
	hostComiled, _ := regexp.Compile(hostPattern)
	hostPatternEntry.hostCompiled = hostComiled
	hostPatternEntry.handlers = hosthandlers
	return hostPatternEntry
}

//Specifies mappings between URLs and handlers.
type UrlSpec struct {
	Name         string
	pattern      string
	HandlerClass HandlerInterface
	Kwargs       Dictionary
	Regexps      *regexp.Regexp
	path         string
	GroupCount   int
}

//	    Parameters:
//        * ``pattern``: Regular expression to be matched.  Any groups
//          in the regex will be passed in to the handler's get/post/etc
//          methods as arguments.
//
//        * ``handlerclass``: `RequestHandler` subclass to be invoked.
//
//        * ``kwargs`` (optional): A dictionary of additional arguments
//          to be passed to the handlerclass's Init.
//
//        * ``name`` (optional): A name for this handlerclass.  Used by
//          `Application.ReverseUrl`.
func NewUrlSpec(pattern string, handlerclass HandlerInterface,
	name string, params Dictionary) UrlSpec {

	urlspec := UrlSpec{}
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}

	urlspec.pattern = pattern
	urlspec.Name = name
	urlspec.HandlerClass = handlerclass
	urlspec.Regexps, _ = regexp.Compile(pattern)
	urlspec.path, urlspec.GroupCount = urlspec.findGroups()
	urlspec.Kwargs = params
	return urlspec

}

func (urls *UrlSpec) Pattern() string {
	return urls.pattern
}

func (urls *UrlSpec) findGroups() (string, int) {
	/*Returns a tuple (reverse string, group count) for a url.
	For example: Given the url pattern /([0-9]{4})/([a-z-]+)/, this method
	would return ('/%s/%s/', 2).
	*/
	pattern := urls.pattern
	if strings.HasPrefix(pattern, "^") {
		pattern = pattern[1:]
	}
	if strings.HasSuffix(pattern, "$") {
		pattern = pattern[:len(pattern)]
	}
	if numsubexp, parenCount := urls.Regexps.NumSubexp(),
		strings.Count(pattern, "("); numsubexp != parenCount {
		// The pattern is too complicated for our simplistic matching,
		// so we can't support reversing it.
		return "", 0
	}

	pieces := []string{}
	for _, franment := range strings.Split(pattern, "(") {
		if strings.Contains(franment, ")") {
			parenLoc := strings.Index(franment, ")")
			if parenLoc >= 0 {
				franment = fmt.Sprintf("%s%s", "%s", franment[parenLoc+1:])
				pieces = append(pieces, franment)
			}

		} else {
			pieces = append(pieces, franment)
		}

	}

	return strings.Join(pieces, ""), urls.Regexps.NumSubexp()
}

func (urls *UrlSpec) Reverse(args ...string) (string, error) {
	if urls.path == "" {
		return "", errors.New(fmt.Sprintf("Cannot reverse url regex %s", urls.pattern))
	}
	if int(len(args)) != urls.GroupCount {
		return "", errors.New("required number of arguments not found")
	}
	if ok := int8(len(args)); ok == 0 {
		return urls.path, nil
	}
	converted_args := []interface{}{}
	for _, arg := range args {
		converted_args = append(converted_args, arg)
	}
	//fmt.Println(urls.path, converted_args)
	return fmt.Sprintf(urls.path, converted_args...), nil
}

type Dictionary map[string]interface{}

func NullDictionary() Dictionary {
	return Dictionary{}
}

func AddRouter(pattern string, handler HandlerInterface, params Dictionary, name string) UrlSpec {
	return NewUrlSpec(pattern, handler, name, params)
}
