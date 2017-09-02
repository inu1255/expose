package expose

import (
	"log"
	"net"
	"os"

	"github.com/inu1255/expose/config"
)

type Connecter struct {
	rpcc *RpcClient
}

// connect to A client with id
func (this *Connecter) Connect(id string) {
	// create conn
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Println(err)
		return
	}
	// get self remote address
	server_addr, _ := net.ResolveUDPAddr("udp", config.Cfg.Client.UdpAddress)
	n, err := conn.WriteTo([]byte("self"), server_addr)
	if err != nil {
		log.Println("can not write to udp server", err)
		return
	}
	b := make([]byte, 32)
	n, _, err = conn.ReadFrom(b)
	if err != nil {
		log.Println("can not read from udp server", err)
		return
	}
	srcAddr := string(b[:n])
	log.Println("got my remote addr", srcAddr)
	// tell A client srcAddr and get A client addr
	addr, err := this.rpcc.askAddr(srcAddr, id)
	if err != nil {
		log.Println("ask for A client address failed", err)
		return
	}
	raddr, _ := net.ResolveUDPAddr("udp", addr)
	// udp forward stdin to A client and A client out to stdout
	NewIoUDP(conn, raddr).Tee(os.Stdin, os.Stdout)
}

func NewConnecter() *Connecter {
	c := new(Connecter)
	c.rpcc = NewRpcClient()
	return c
}
