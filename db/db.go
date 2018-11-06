package db

import (
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
)

//sql engine
var Engine *xorm.Engine
var Client *redis.Client

//starts sql engine
//every table must have to be created in db manually

func StartEngine() {
	var err error
	Engine, err = xorm.NewEngine("mysql", "root:@/online-judge?charset=utf8")

	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	Engine.SetDefaultCacher(cacher)

	if err != nil {
		panic(err.Error())
	}

	Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
