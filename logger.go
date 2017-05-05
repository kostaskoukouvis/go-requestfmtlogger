package logger

import (
	"context"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/go-stack/stack"
)

// LoggerConfig contains a boolean that judges whether the
// logger should write in syslog or not.
type LoggerConfig struct {
	SysWrite bool
}

type contextKey string

type loggerItem struct {
	sync.RWMutex
	m      map[string]interface{}
	url    string
	method string
}

var (
	logCtx = contextKey("reqlogger")
	mlog   *log.Logger
)

func init() {
	mlog = log.New(os.Stdout, "", 0)
}

func (c *LoggerConfig) RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		i := newLoggerItem(r)
		ww := &responseWriterWrapper{0, 0, w}
		t1 := time.Now()
		ctx := context.WithValue(r.Context(), logCtx, i)
		next.ServeHTTP(ww, r.WithContext(ctx))

		t2 := time.Now()
		finilizeRequestLogger(i, c.SysWrite, ww.Status(), ww.Size(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func newLoggerItem(r *http.Request) *loggerItem {
	i := &loggerItem{}
	i.url = r.URL.String()
	i.method = r.Method
	i.m = make(map[string]interface{})
	return i
}

func Log(r *http.Request, values ...interface{}) {
	i, _ := r.Context().Value(logCtx).(*loggerItem)
	i.Lock()
	dataMap := mapify(values)
	for k, v := range dataMap {
		i.m[k] = v
	}
	i.m["file"] = stack.Caller(1)
	i.Unlock()
}

func finilizeRequestLogger(i *loggerItem, sysWrite bool, status int, size int, elapsed time.Duration) {
	i.Lock()
	rMap := make(map[string]interface{})
	rMap["method"] = i.method
	rMap["url"] = i.url
	rMap["status"] = status
	rMap["size"] = size
	rMap["duration"] = elapsed
	// print stuff out
	if sysWrite {
		byteMessage := formatSyslog(rMap, i.m)
		// check status to assume priority
		var prio syslog.Priority
		switch {
		case status < 400:
			prio = syslog.LOG_INFO
		case status < 500:
			prio = syslog.LOG_WARNING
		default:
			prio = syslog.LOG_ERR
		}
		w, _ := syslog.New(prio, "")
		w.Write(byteMessage)
		w.Close()
	} else {
		message := formatTerminal(rMap, i.m)
		mlog.Printf("%s", message)
	}
	i.Unlock()
}

func mapify(arr []interface{}) map[string]interface{} {
	if len(arr)%2 != 0 {
		// If the given values aren't even remove the last one
		arr = append(arr[:len(arr)-1], arr[len(arr):]...)
	}
	result := make(map[string]interface{})
	// work the array in pairs.
	for i := 0; i < len(arr); i += 2 {
		// unless the first value is a string, dump the pair.
		v := reflect.TypeOf(arr[i]).Kind()
		if v == reflect.String {
			result[arr[i].(string)] = arr[i+1]
		}
	}
	return result
}
