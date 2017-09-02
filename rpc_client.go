package expose

import (
	"context"
	"log"

	"github.com/inu1255/expose/config"
	"github.com/inu1255/expose/msg"
	"google.golang.org/grpc"
)

type RpcClient struct {
	conn         *grpc.ClientConn
	ctx          context.Context
	ComputerServ msg.ComputerServiceClient
}

func (this *RpcClient) RegistComputer() (*msg.Computer, error) {
	body := &msg.ComputerRegisterBody{Name: "xxx", Mac: "ac:bc:32:98:b9:37"}
	computer, err := this.ComputerServ.Register(this.ctx, body)
	return computer, err
}

func (this *RpcClient) askAddr(srcAddr, id string) (string, error) {
	body := &msg.SrcDst{srcAddr, id}
	addr, err := this.ComputerServ.AskAddr(this.ctx, body)
	if err != nil {
		return "", err
	}
	return addr.GetS(), nil
}

func NewRpcClient() *RpcClient {
	conn, err := grpc.Dial(config.Cfg.Client.RpcAddress, grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		return nil
	}
	rpcc := new(RpcClient)
	rpcc.conn = conn
	rpcc.ctx = context.Background()
	rpcc.ComputerServ = msg.NewComputerServiceClient(conn)
	return rpcc
}
