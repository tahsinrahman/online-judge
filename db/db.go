package db

import (
	"github.com/go-xorm/xorm"
)

//sql engine
var Engine *xorm.Engine

//starts sql engine
//every table must have to be created in db manually

func StartEngine() {
	var err error
	Engine, err = xorm.NewEngine("mysql", "root:@/online-judge?charset=utf8")

	if err != nil {
		//TODO: response 500 internal server error
		panic(err.Error())
	}
	//Engine.SetMapper(core.NewCacheMapper(new(core.SameMapper)))
}
