package service

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/inu1255/expose/config"
	"github.com/inu1255/expose/msg"
)

type Addr int8

type OnlineServer struct {
	conns *sync.Map // id -> net.Conn id对应的tcp连接
	addrs *sync.Map // id -> chan string  id对应的地址
}

func (this *OnlineServer) Run() {
	tcp_addr, _ := net.ResolveTCPAddr("tcp", config.Cfg.Server.TcpAddress)
	log.Println("listen tcp:", tcp_addr)
	lis, err := net.Listen("tcp", tcp_addr.String())
	if err != nil {
		log.Println(err)
		return
	}
	go this.startUDP()
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err, conn)
			continue
		}
		go this.keepAlive(conn)
	}
}

// keep A client connect
func (this *OnlineServer) keepAlive(conn net.Conn) {
	b := make([]byte, 32)
	n, err := conn.Read(b)
	if err != nil || n != 24 {
		log.Println(n, string(b))
		conn.Close()
		return
	}
	id := string(b[:n])
	if _, err := ComputerServ.Exist(nil, &msg.String{id}); err != nil {
		log.Println(id)
		conn.Close()
		return
	}
	this.conns.Store(id, conn)
}

// start udp server
// two usage:
// 1.ask for self remote address
// 	method: send something not $id to the server
// 	the response is your remote address
// 2.tell your remote address to client
// 	method: send your $id to the server
// 	the server will find the asker by $id and tell him your remote address by channel
func (this *OnlineServer) startUDP() {
	udp_addr, _ := net.ResolveUDPAddr("udp", config.Cfg.Server.UdpAddress)
	log.Println("listen udp:", udp_addr)
	conn, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		log.Println(err)
		return
	}
	b := make([]byte, 32)
	for {
		// A client's address
		n, addr, err := conn.ReadFrom(b)
		if err != nil {
			log.Println(addr, err)
			continue
		}
		id := string(b[:n])
		// find the asker and tell him your remote address
		if v, ok := this.addrs.Load(id); ok {
			log.Println(id, "is at", addr)
			c := v.(chan string)
			c <- addr.String()
		} else {
			log.Println("ask self address at", addr)
			conn.WriteTo([]byte(addr.String()), addr)
		}
	}
}

// get the A client's address by id
func (this *OnlineServer) AskAddr(srcAddr, dst string) (string, error) {
	if v, ok := this.conns.Load(dst); ok {
		conn := v.(net.Conn)
		if _, err := conn.Write([]byte(srcAddr + "\n")); err != nil {
			log.Println("ask addr failed:", err)
			this.conns.Delete(dst)
			conn.Close()
			return "", err
		}
		c := make(chan string)
		this.addrs.Store(dst, c)
		var s string
		t := time.NewTimer(time.Duration(config.Cfg.Server.AskAddrTimeOut) * time.Second)
		select {
		case s = <-c:
		case <-t.C:
			return "", TimeOutError
		}
		return s, nil
	}
	return "", NotExistError
}

var OnlineServ = &OnlineServer{&sync.Map{}, &sync.Map{}}
