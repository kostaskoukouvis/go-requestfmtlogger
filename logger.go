package logger

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type contextKey string

type logEntry struct {
	sync.Mutex
	m map[*http.Request]map[string]interface{}
}

const (
	loggerCtx contextKey = "request_logger"
)

var (
	entry logEntry
)

func RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		logEntry := initRequestLogger(r)
		ww := &ResponseWriterWrapper{0, 0, w}
		t1 := time.Now()
		log.Print(logEntry)

		next.ServeHTTP(ww, withLogEntry(r, logEntry))

		t2 := time.Now()
		finilizeRequestLogger(r, ww.Status(), ww.Size(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func withLogEntry(r *http.Request, e logEntry) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), loggerCtx, e))
	return r
}

func getLogEntry(r *http.Request) logEntry {
	e, _ := r.Context().Value(loggerCtx).(logEntry)
	return e
}

func initRequestLogger(r *http.Request) logEntry {
	log.Print("initReqLog")
	entry.m = make(map[*http.Request]map[string]interface{})
	entry.m[r] = make(map[string]interface{})
	entry.m[r]["url"] = r.URL.String()
	entry.m[r]["method"] = r.Method
	log.Printf("%+v", entry.m)
	return entry
}

func Log(r *http.Request, values ...interface{}) error {
	return nil
}

func finilizeRequestLogger(r *http.Request, status int, size int, elapsed time.Duration) {
	log.Print(r.Context())
	log.Print(entry.m, status, elapsed)
}
