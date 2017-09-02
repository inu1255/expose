package expose

import (
	"bufio"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/inu1255/expose/config"
)

type Client struct {
	id string
}

func (this *Client) Run() error {
	conn, err := net.Dial("tcp", config.Cfg.Client.TcpAddress)
	if err != nil {
		log.Println(err)
		return err
	}
	// tell server your id
	n, err := conn.Write([]byte(this.id))
	if err != nil {
		log.Println(n, err)
		return err
	}
	log.Println("server start,wait connect")
	r := bufio.NewReader(conn)
	for {
		addr, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println(this.id, "is not registered")
				return err
			}
			log.Println(n, err)
			continue
		}
		addr = strings.Trim(addr, " \t\n")
		log.Println("new connect from", addr)
		go this.handleConnect(addr)
	}
}

// start udp, wait B client connect
func (this *Client) handleConnect(addr string) {
	client_addr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Println(err)
		return
	}
	server_addr, _ := net.ResolveUDPAddr("udp", config.Cfg.Client.UdpAddress)
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	// tell B client your remote address by send your id to server
	n, err := conn.WriteTo([]byte(this.id), server_addr)
	if err != nil {
		log.Println(n, err)
		return
	}
	ss := strings.Fields(config.Cfg.Client.Command)
	cmd := exec.Command(ss[0], ss[1:]...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return
	}
	wr := NewIoUDP(conn, client_addr)
	WaitAny(func(c chan bool) {
		ForwardUDP(wr, stdin, time.Duration(config.Cfg.Client.ReadTimeOut)*time.Second)
		c <- true
	}, func(c chan bool) {
		Forward(stdout, wr)
		c <- true
	}, func(c chan bool) {
		Forward(stderr, wr)
		c <- true
	})
	cmd.Process.Kill()
	log.Println(addr, "disconnect")
}

func NewClient(id string) *Client {
	return &Client{id}
}

func WaitAny(funcs ...func(chan bool)) {
	count := len(funcs)
	c := make(chan bool)
	for i := 0; i < count; i++ {
		fn := funcs[i]
		go fn(c)
	}
	<-c
}

func WaitAll(funcs ...func(chan bool)) {
	count := len(funcs)
	c := make(chan bool)
	for i := 0; i < count; i++ {
		fn := funcs[i]
		fn(c)
	}
	for i := 0; i < count; i++ {
		<-c
	}
}
