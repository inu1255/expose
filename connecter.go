package expose

import (
	"log"
	"net"
	"os"

	"github.com/inu1255/expose/config"
	"github.com/inu1255/expose/msg"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Connecter struct {
	conn         *grpc.ClientConn
	ctx          context.Context
	ComputerServ msg.ComputerServiceClient
}

func (this *Connecter) RegistComputer() (*msg.Computer, error) {
	body := &msg.ComputerRegisterBody{Name: "xxx", Mac: "ac:bc:32:98:b9:37"}
	computer, err := this.ComputerServ.Register(this.ctx, body)
	return computer, err
}

func (this *Connecter) askAddr(srcAddr, id string) (string, error) {
	body := &msg.SrcDst{srcAddr, id}
	addr, err := this.ComputerServ.AskAddr(this.ctx, body)
	if err != nil {
		return "", err
	}
	return addr.GetS(), nil
}

// connect to A client with id
func (this *Connecter) Connect(id string) {
	// create conn
	conn, err := net.ListenUDP("udp4", nil)
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
	addr, err := this.askAddr(srcAddr, id)
	if err != nil {
		log.Println("ask for A client address failed", err)
		return
	}
	raddr, _ := net.ResolveUDPAddr("udp", addr)
	// udp forward stdin to A client and A client out to stdout
	NewIoUDP(conn, raddr).Tee(os.Stdin, os.Stdout)
}

func NewConnecter() *Connecter {
	conn, err := grpc.Dial(config.Cfg.Client.RpcAddress, grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		return nil
	}
	connecter := new(Connecter)
	connecter.conn = conn
	connecter.ctx = context.Background()
	connecter.ComputerServ = msg.NewComputerServiceClient(conn)
	return connecter
}
