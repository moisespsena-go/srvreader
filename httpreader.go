package srvreader

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type HTTPServerReader struct {
	KeepAlivePeriod time.Duration
	proto, addr     string
	s               *http.Server
	l               net.Listener
	w               io.Writer
}

func NewHTTPServerReader(addr string, w io.Writer) *HTTPServerReader {
	proto, addr := parseAddr(addr)
	proto = "tcp" + strings.TrimPrefix(proto, "http")
	return &HTTPServerReader{proto: proto, addr: addr, w: w}
}

func (this *HTTPServerReader) Close() error {
	this.s.Shutdown(context.Background())
	return nil
}

func (this *HTTPServerReader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		for {
			_, b, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if b[len(b)] != '\n' {
				b = append(b, '\n')
			}
			if _, err := this.w.Write(b); err != nil {
				return
			}

		}
	} else if r.Method == http.MethodPost {
		if _, err := io.Copy(this.w, &lfreader{r: r.Body}); err != nil {
			log.Error("HTTP request body copy failed: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "bad http method", http.StatusBadRequest)
	}
}

func (this *HTTPServerReader) ListenAndServe() (err error) {
	if this.l, err = net.Listen(this.proto, this.addr); err != nil {
		return
	}
	this.s = &http.Server{Handler: this}
	return this.s.Serve(NewTCPKeepAliveListener(this.l.(*net.TCPListener), this.KeepAlivePeriod))
}
