package expose

import (
	"log"
	"net"

	"github.com/inu1255/expose/config"
	"github.com/inu1255/expose/msg"
	"github.com/inu1255/expose/service"
	"google.golang.org/grpc"
)

func RpcServer() {
	lis, err := net.Listen("tcp", config.Cfg.Server.RpcAddress)
	if err != nil {
		log.Println(err)
		return
	}
	s := grpc.NewServer()
	msg.RegisterComputerServiceServer(s, service.ComputerServ)
	s.Serve(lis)
}

func RunServer() {
	go RpcServer()
	service.OnlineServ.Run()
}
