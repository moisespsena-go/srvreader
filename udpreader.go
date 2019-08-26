package srvreader

import (
	"io"
	"net"
)

type UDPServerReader struct {
	proto, address string
	w              io.Writer
	buffer         []byte
	pc             net.PacketConn
}

func (this *UDPServerReader) Close() error {
	if this.pc != nil {
		return this.pc.Close()
	}
	return nil
}

func NewUDPServer(addr string, maxBufferSize int16, w io.Writer) *UDPServerReader {
	proto, addr := parseAddr(addr)
	return &UDPServerReader{proto: proto, address: addr, w: w, buffer: make([]byte, maxBufferSize)}
}

func (this *UDPServerReader) ListenAndServe() (err error) {
	this.pc, err = net.ListenPacket(this.proto, this.address)
	if err != nil {
		return
	}

	defer this.Close()

	var n int

	for {
		if n, _, err = this.pc.ReadFrom(this.buffer); err != nil {
			return
		}

		if n > 0 {
			b := this.buffer[0:n]
			if b[n-1] != '\n' {
				b = append(b, '\n')
			}
			if _, err = this.w.Write(b); err != nil {
				return
			}
		}
	}
}
