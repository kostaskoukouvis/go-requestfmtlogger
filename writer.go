package logger

import (
	"net/http"
)

// ResponseWriterWrapper wraps around the http response writer.
type responseWriterWrapper struct {
	status int
	size   int
	http.ResponseWriter
}

// Status returns the status code of the response.
func (wr *responseWriterWrapper) Status() int {
	return wr.status
}

// Size returns the size of the body of the response.
func (wr *responseWriterWrapper) Size() int {
	return wr.size
}

// Header returns the http.Header of the wrapped response writer.
func (wr *responseWriterWrapper) Header() http.Header {
	return wr.ResponseWriter.Header()
}

// Write captures the size and forwards it to the wrapper.
func (wr *responseWriterWrapper) Write(b []byte) (int, error) {
	if wr.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		wr.status = http.StatusOK
	}
	size, err := wr.ResponseWriter.Write(b)
	wr.size += size
	return size, err
}

// WriteHeader captures the status and forwards it to the wrapper.
func (wr *responseWriterWrapper) WriteHeader(status int) {
	wr.status = status
	wr.ResponseWriter.WriteHeader(status)
}

// // Hijack makes the responseWriterWrapper conform to the hijacking interface
// func (wr *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
// 	hj, ok := wr.ResponseWriter.(http.Hijacker)
// 	if !ok {
// 		return nil, nil, fmt.Errorf("ResponseWriter does not support Hijack")
// 	}
// 	return hj.Hijack()
// }
