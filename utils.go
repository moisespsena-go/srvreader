package srvreader

import (
	"io"
	"net"
	"strings"
	"time"
	"unicode"
)

type lfreader struct {
	r    io.Reader
	done bool
}

func (this lfreader) Read(p []byte) (n int, err error) {
	if this.done {
		err = io.EOF
	} else if n, err = this.r.Read(p); err == io.EOF {
		if n < len(p) {
			p[n] = '\n'
			n++
		} else {
			p[0] = '\n'
			n = 1
			err = nil
		}
		this.done = true
	}
	return
}

func parseAddr(address string) (network, addr string) {
	pos := strings.IndexRune(address, ':')
	network, addr = address[0:pos], address[pos+1:]
	if !strings.HasSuffix(network, "6") && strings.IndexRune(addr, '[') >= 0 {
		network += "6"
	}
	return
}

func IsProto(address, proto string) bool {
	if pos := strings.IndexRune(address, ':'); pos != -1 {
		for unicode.IsDigit(rune(address[pos-1])) {
			pos--
		}
		part := address[0:pos]
		return proto == part
	}
	return false
}

// TCPKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type TCPKeepAliveListener struct {
	*net.TCPListener
	Period time.Duration
}

func NewTCPKeepAliveListener(TCPListener *net.TCPListener, Period ...time.Duration) *TCPKeepAliveListener {
	if len(Period) == 0 || Period[0] == 0 {
		Period = []time.Duration{3 * time.Minute}
	}
	return &TCPKeepAliveListener{TCPListener: TCPListener, Period: Period[0]}
}

func (this TCPKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := this.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(this.Period)
	return tc, nil
}
