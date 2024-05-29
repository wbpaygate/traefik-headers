package traefik_headers

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
)

func (h *Headers) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.next.ServeHTTP(&responseWriter{
		rw:      rw,
		headers: ghs.headers[int(atomic.LoadInt32(ghs.curheader))],
	}, req)
}

type responseWriter struct {
	rw      http.ResponseWriter
	headers *headers
}

func (r *responseWriter) Header() http.Header {
	return r.rw.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	return r.rw.Write(bytes)
}

func (r *responseWriter) WriteHeader(code int) {
	head := r.rw.Header()
	for k, vv := range r.headers.headers {
		if _, ok := head[k]; !ok {
			for _, v := range vv {
				r.rw.Header().Add(k, v)
			}
		}
	}
	r.rw.WriteHeader(code)
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.rw.(http.Hijacker)
	if !ok || hj == nil {
		return nil, nil, fmt.Errorf("http.Hijacker interface is not implemented in given response writer")
	}

	return hj.Hijack()
}
