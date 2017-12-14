package main

import (
	"fmt"

	"github.com/go-xorm/xorm"
	_ "github.com/ziutek/mymysql/mysql"
	"gopkg.in/macaron.v1"
)

var engine *xorm.Engine

func main() {
	var err error
	engine, err = xorm.NewEngine("mymysql", "root@/online-judge?charset=utf8")

	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	} else {
		fmt.Println(engine)
		engine.DBMetas()
	}

	m := macaron.Classic()
	m.Use(macaron.Renderer())
	//m.Use(session.Sessioner())

	m.Get("/", func(ctx *macaron.Context) {
		//		sess.Set("session", "session middleware")
		//		ctx.Data["Login"] = sess.Get("session").(string)
		ctx.HTML(200, "index")
	})

	m.Get("/signup", func(ctx *macaron.Context) {
		ctx.HTML(200, "signup")
	})
	m.Post("/signup", func(ctx *macaron.Context) {
		//insert to DB
		ctx.Redirect(ctx.Req.URL.Host, 301)
	})

	m.Get("/signin", func(ctx *macaron.Context) {
		ctx.HTML(200, "signin")
	})
	m.Post("/signin", func(ctx *macaron.Context) {
		//insert to DB
		ctx.Redirect(ctx.Req.URL.Host, 301)
	})

	m.Get("/signout", func(ctx *macaron.Context) {
		ctx.HTML(200, "signout")
	})

	m.Run()
}
