package main

import (
	"log"
	"os"

	"github.com/inu1255/expose"
	"github.com/inu1255/expose/config"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app     = kingpin.New("expose", "a tool expose inner network")
	server  = app.Command("server", "start a outer network service for expose client")
	client  = app.Command("client", "start a inner network service to expose yourself to others")
	connect = app.Command("connect", "connect a exposed client")
	id      = connect.Arg("id", "the client id you want to connect").Required().String()
)

func main() {
	// log.SetFlags(log.Ltime | log.Llongfile)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case server.FullCommand():
		expose.RunServer()
	case client.FullCommand():
		if config.Cfg.Client.Id == "" {
			log.Println("you don't have a id register...")
			rpcc := expose.NewRpcClient()
			computer, err := rpcc.RegistComputer()
			if err != nil {
				log.Println(err)
				return
			}
			config.Cfg.Client.Id = computer.GetId()
			log.Println("your id is", computer.GetId())
			err = config.Save()
			if err != nil {
				log.Println("save failed", err)
			}
		}
		server := expose.NewClient(config.Cfg.Client.Id)
		server.Run()
	case connect.FullCommand():
		client := expose.NewConnecter()
		client.Connect(*id)
	}
}
