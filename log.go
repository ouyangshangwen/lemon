package lemon

import (
	"log"
	"os"
)

type LemonLog struct {
	Log *log.Logger
}

func NewLog() *LemonLog {
	Log := log.New(os.Stderr, "[Debug]", log.LstdFlags)
	return &LemonLog{Log: Log}
}

func (llog *LemonLog) Debug(message interface{}) {
	llog.Log.SetPrefix("[Debug] ")
	llog.Log.Println(message)
}

func (llog *LemonLog) Info(message interface{}) {
	llog.Log.SetPrefix("[Info] ")
	llog.Log.Println(message)
}

func (llog *LemonLog) Warning(message interface{}) {
	llog.Log.SetPrefix("[Warning] ")
	llog.Log.SetFlags(llog.Log.Flags() | log.Llongfile)
	llog.Log.Println(message)
}

func (llog *LemonLog) Trace(message interface{}) {
	llog.Log.SetPrefix("[Trace] ")
	llog.Log.SetFlags(llog.Log.Flags() | log.Llongfile)
	llog.Log.Println(message)
}

func (llog *LemonLog) Error(message interface{}) {
	llog.Log.SetPrefix("[Error] ")
	llog.Log.SetFlags(llog.Log.Flags() | log.Llongfile)
	llog.Log.Println(message)
}

func (llog *LemonLog) Fatal(message interface{}) {
	llog.Log.SetPrefix("[Fatal] ")
	llog.Log.SetFlags(llog.Log.Flags() | log.Llongfile)
	llog.Log.Println(message)
}

var lemonLag *LemonLog = NewLog()
