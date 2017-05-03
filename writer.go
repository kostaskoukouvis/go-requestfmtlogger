package logger

import (
	"net/http"
)

// ResponseWriterWrapper wraps around the http response writer.
type ResponseWriterWrapper struct {
	status int
	size   int
	http.ResponseWriter
}

// Status returns the status code of the response.
func (wr *ResponseWriterWrapper) Status() int {
	return wr.status
}

// Size returns the size of the body of the response.
func (wr *ResponseWriterWrapper) Size() int {
	return wr.size
}

// Header returns the http.Header of the wrapped response writer.
func (wr *ResponseWriterWrapper) Header() http.Header {
	return wr.ResponseWriter.Header()
}

// Write captures the size and forwards it to the wrapper.
func (wr *ResponseWriterWrapper) Write(b []byte) (int, error) {
	if wr.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		wr.status = http.StatusOK
	}
	size, err := wr.ResponseWriter.Write(b)
	wr.size += size
	return size, err
}

// WriteHeader captures the status and forwards it to the wrapper.
func (wr *ResponseWriterWrapper) WriteHeader(status int) {
	wr.status = status
	wr.ResponseWriter.WriteHeader(status)
}

// // Hijack makes the ResponseWriterWrapper conform to the hijacking interface
// func (wr *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
// 	hj, ok := wr.ResponseWriter.(http.Hijacker)
// 	if !ok {
// 		return nil, nil, fmt.Errorf("ResponseWriter does not support Hijack")
// 	}
// 	return hj.Hijack()
// }
