package logger

import (
	"log"
	"log/syslog"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/go-stack/stack"
)

type LoggerConfig struct {
	SysWrite bool
}

var (
	mlog  *log.Logger
	mutex sync.RWMutex
	data  = make(map[*http.Request]map[string]interface{})
)

func init() {
	mlog = log.New(os.Stdout, "", 0)
}

func (c *LoggerConfig) RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		initRequestLogger(r)
		ww := &responseWriterWrapper{0, 0, w}
		t1 := time.Now()

		next.ServeHTTP(ww, r)

		t2 := time.Now()
		finilizeRequestLogger(r, c.SysWrite, ww.Status(), ww.Size(), t2.Sub(t1))
		clear(r)
	}
	return http.HandlerFunc(fn)
}

func initRequestLogger(r *http.Request) {
	mutex.Lock()
	data[r] = make(map[string]interface{})
	mutex.Unlock()
}

func Log(r *http.Request, values ...interface{}) {
	mutex.Lock()
	if data[r] == nil {
		// if the log wasn't initialized, burn it all to hell...
		return
	}
	dataMap := mapify(values)
	for k, v := range dataMap {
		data[r][k] = v
	}
	data[r]["file"] = stack.Caller(1)
	mutex.Unlock()
}

func finilizeRequestLogger(r *http.Request, sysWrite bool, status int, size int, elapsed time.Duration) {
	mutex.Lock()
	if data[r] == nil {
		// if the log wasn't initialized, burn it all to hell...
		return
	}

	rMap := make(map[string]interface{})
	rMap["method"] = r.Method
	rMap["url"] = r.URL.String()
	rMap["status"] = status
	rMap["size"] = size
	rMap["duration"] = elapsed
	// print stuff out
	if sysWrite {
		byteMessage := formatSyslog(rMap, data[r])
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
		message := formatTerminal(rMap, data[r])
		mlog.Printf("%s", message)
	}
	mutex.Unlock()
}

func clear(r *http.Request) {
	mutex.Lock()
	delete(data, r)
	mutex.Unlock()
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
