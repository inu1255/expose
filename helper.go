package expose

import (
	"io"
	"log"
	"net"
	"reflect"
	"time"
)

// forward r --> w
func Forward(r io.Reader, w io.Writer) {
	b := make([]byte, 1024)
	for {
		n, err := r.Read(b)
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
		_, err = w.Write(b[:n])
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
	}
}

// forward r --> w with timeout
func ForwardUDP(r *ioUDP, w io.Writer, timeout time.Duration) {
	b := make([]byte, 1024)
	for {
		if timeout > 0 {
			r.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err := r.Read(b)
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
		_, err = w.Write(b[:n])
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
	}
}

// UDPConn --> fat interface io.WriteReader
type ioUDP struct {
	*net.UDPConn
	addr net.Addr
}

func (this *ioUDP) Read(b []byte) (int, error) {
	n, addr, err := this.UDPConn.ReadFrom(b)
	if err == nil && addr.String() != this.addr.String() {
		return 0, nil
	}
	return n, err
}

func (this *ioUDP) Write(b []byte) (int, error) {
	return this.UDPConn.WriteTo(b, this.addr)
}

// in --> conn
// conn --> out
func (conn *ioUDP) Tee(in io.Reader, out io.Writer) {
	go Forward(in, conn)
	Forward(conn, out)
}

func NewIoUDP(conn *net.UDPConn, addr net.Addr) *ioUDP {
	return &ioUDP{conn, addr}
}
