package service

import (
	"log"

	"github.com/inu1255/expose/msg"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Computer Entity
type ComputerService struct {
}

func (this *ComputerService) Register(ctx context.Context, in *msg.ComputerRegisterBody) (*msg.Computer, error) {
	c := new(msg.Computer)
	c.Mac = in.GetMac()
	c.Name = in.GetName()
	if err := db.C("computer").Find(bson.M{"mac": c.Mac, "name": c.Name}).One(c); err == mgo.ErrNotFound {
		c.Id = bson.NewObjectId()
		err = db.C("computer").Insert(c)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return c, nil
}

func (this *ComputerService) AskAddr(ctx context.Context, in *msg.SrcDst) (*msg.String, error) {
	addr, err := OnlineServ.AskAddr(in.GetSrc(), in.GetDst())
	if err != nil {
		return nil, err
	}
	return &msg.String{addr}, nil
}

func (this *ComputerService) Exist(ctx context.Context, in *msg.String) (*msg.Int, error) {
	n, err := db.C("computer").FindId(bson.ObjectIdHex(in.GetS())).Count()
	if err != nil {
		log.Println(n, err)
		return nil, err
	}
	if n < 1 {
		log.Println(n, err)
		return nil, NotExistError
	}
	return &msg.Int{1}, nil
}

var ComputerServ msg.ComputerServiceServer = new(ComputerService)

/*****************************************************************************
 *                                 api above                                 *
 *****************************************************************************/
