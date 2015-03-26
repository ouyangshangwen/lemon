package lemon

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"time"
)

type Lemon struct {
	Server  *http.Server
	app     *Application
	address string
	port    int
	network string //used in the mode of fastcgi, value is ''(standard I/O), 'tcp', 'unix'
}

func NewLemon() *Lemon {
	return &Lemon{Server: &http.Server{}}
}

func (lem *Lemon) Instance(urlSpecs []UrlSpec, settings map[string]interface{}) *Lemon {

	app := &Application{}
	lem.app = app
	lem.Server.Handler = app
	lem.app.Init(urlSpecs, settings)
	//addr := app.HttpAddr
	var addr string
	//addr = fmt.Sprintf("%s:%d", app.HttpAddr, app.Port)

	lem.Server.Addr = addr
	lem.Server.ReadTimeout = app.ReadTimeOut
	lem.Server.WriteTimeout = app.WriteTimeOut
	return lem
}

func (lem *Lemon) Listen(address string, port int) {

	lem.port = port
	lem.address = address

}

func (lem *Lemon) FCGILoop(network string) {
	if network == "" {
		err := fcgi.Serve(nil, lem.Server.Handler) // standard I/O
		if err == nil {
			fmt.Println("Use FCGI via standard I/O")
		} else {
			fmt.Println("Cannot use FCGI via standard I/O", err)
		}
	} else if network == "tcp" {
		lem.address = fmt.Sprintf("%s:%d", lem.address, lem.port)
	}
	l, err := net.Listen(lem.network, lem.address)
	if err != nil {
		fmt.Println("Listen: ", err)
	}
	err = fcgi.Serve(l, lem.Server.Handler)
	if err != nil {
		fmt.Println("Listen Error: ", err)
	}

}

func (lem *Lemon) Loop() {
	lem.Server.Addr = fmt.Sprintf("%s:%d", lem.address, lem.port)

	if len(lem.app.CertFile) != 0 || len(lem.app.KeyFile) != 0 {
		_, HasCertFile := os.Stat(lem.app.CertFile)
		if HasCertFile != nil {
			errorString := fmt.Sprintf("certfile %s does not exist", lem.app.CertFile)
			errors.New(errorString)
			return

		}
		_, HasKeyFile := os.Stat(lem.app.KeyFile)
		if HasKeyFile != nil {
			errorString := fmt.Sprintf("keyfile %s does not exist", lem.app.KeyFile)
			errors.New(errorString)
			return
		}
		lem.ListenHttpTLs()
	} else {
		lem.ListenHttp()
	}

}

func (lem *Lemon) ListenHttpTLs() {
	time.Sleep(20 * time.Microsecond)
	fmt.Println("https server Running on ", lem.address)
	err := lem.Server.ListenAndServeTLS(lem.app.CertFile, lem.app.KeyFile)
	if err != nil {
		fmt.Println("ListenAndServeTLS: ", err)
		time.Sleep(100 * time.Microsecond)
	}
}

func (lem *Lemon) ListenHttp() {
	time.Sleep(20 * time.Microsecond)
	fmt.Println("http server Running on ", lem.Server.Addr)
	err := lem.Server.ListenAndServe()
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
		time.Sleep(100 * time.Microsecond)
	}
}
