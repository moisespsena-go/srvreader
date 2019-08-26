package srvreader

import (
	"io"
	"net"
	"time"
)

type TCPServerReader struct {
	KeepAlivePeriod time.Duration
	proto, addr     string
	l               net.Listener
	w               io.Writer
}

func (this *TCPServerReader) Close() error {
	return this.l.Close()
}

func NewTCPServerReader(addr string, w io.Writer) *TCPServerReader {
	proto, addr := parseAddr(addr)
	return &TCPServerReader{proto: proto, addr: addr, w: w}
}

func (this *TCPServerReader) ListenAndServe() (err error) {
	if this.l, err = net.Listen(this.proto, this.addr); err != nil {
		return
	}

	this.l = NewTCPKeepAliveListener(this.l.(*net.TCPListener), this.KeepAlivePeriod)

	var c net.Conn
	for {
		if c, err = this.l.Accept(); err != nil {
			return
		}
		go func(c net.Conn) {
			if _, err := io.Copy(this.w, c); err != nil && err != io.EOF {
				log.Errorf("TCP copy failed: %s", err.Error())
			}
		}(c)
	}
}
