package service

import (
	"log"

	"github.com/inu1255/expose/config"
	"gopkg.in/mgo.v2"
)

var (
	db = NewDb()
)

func NewDb() *mgo.Database {
	sess, err := mgo.Dial(config.Cfg.MongoUrl)
	if err != nil {
		log.Panicln(err)
	}
	sess.SetMode(mgo.Monotonic, true)
	return sess.DB("expose")
}
