package config

import ini "gopkg.in/ini.v1"

var (
	cfg, _ = ini.LooseLoad("expose.ini")
	Cfg    = &config{
		MongoUrl: "127.0.0.1",
		Server: server{
			RpcAddress:     "0.0.0.0:9182",
			TcpAddress:     "0.0.0.0:9183",
			UdpAddress:     "0.0.0.0:9184",
			AskAddrTimeOut: 30,
		},
		Client: client{
			Id:          "",
			ReadTimeOut: 300,
			Command:     "/bin/bash",
			RpcAddress:  "inu1255.cn:9182",
			TcpAddress:  "inu1255.cn:9183",
			UdpAddress:  "inu1255.cn:9184",
		},
	}
)

type config struct {
	MongoUrl string
	Server   server
	Client   client
}

type server struct {
	RpcAddress     string
	TcpAddress     string
	UdpAddress     string
	AskAddrTimeOut int
}

type client struct {
	Id          string
	ReadTimeOut int
	Command     string
	RpcAddress  string
	TcpAddress  string
	UdpAddress  string
}

func init() {
	cfg.MapTo(Cfg)
}

func Save() error {
	err := ini.ReflectFrom(cfg, Cfg)
	if err != nil {
		return err
	}
	return cfg.SaveTo("expose.ini")
}
