package main

import (
	"fmt"

	"github.com/go-macaron/binding"
	"github.com/go-macaron/session"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
)

type Users struct {
	Name     string `form:"name"`
	Username string `form:"username"`
	Password string `form:"password"`
}

var engine *xorm.Engine

//starts an sql engine
//user table is created in DB, not here
func initEngine() {
	var err error
	engine, err = xorm.NewEngine("mysql", "root:@/online-judge?charset=utf8")

	if err != nil {
		panic(err.Error())
	}
}

//get username from cookie, check if username exists
//TODO: securecookie using token
func checkAuthenticationHandler(ctx *macaron.Context, sess session.Store) {
	//	cookie := ctx.GetCookie("user")

	res := sess.Get("user")

	if res == nil {
		return
	}

	//	fmt.Println(&username)

	//user if logged in
	//used it for showing logout option in html
	ctx.Data["Login"] = 1
	ctx.Data["Username"] = res.(string)
	fmt.Println(ctx.Data["Username"])
}

func main() {
	initEngine()

	m := macaron.Classic()
	m.Use(macaron.Renderer()) //for html rendering
	m.Use(session.Sessioner())
	m.Use(checkAuthenticationHandler)

	m.Group("", func() {

		//homepage
		m.Get("/", func(ctx *macaron.Context) {
			ctx.HTML(200, "index")
		})

		//signup form
		m.Get("/signup", func(ctx *macaron.Context) {
			//if logged in, redirect to home
			if ctx.Data["Login"] == 1 {
				ctx.Redirect(ctx.Req.URL.Host, 301)
				return
			}

			ctx.HTML(200, "signup")
		})

		//create new user
		m.Post("/signup", binding.Bind(Users{}), func(ctx *macaron.Context, user Users) {
			//if logged in, redirect to home
			if ctx.Data["Login"] == 1 {
				ctx.Redirect(ctx.Req.URL.Host, 301)
				return
			}
			//insert to DB
			//TODO: encryption for pass

			//check if username already exists
			//if yes, then response with "username exists"
			//else insert new user into db

			//check if exists
			var tmpUser = Users{Username: user.Username}
			has, err := engine.Get(&tmpUser)

			if err != nil || has {
				ctx.Redirect(ctx.Req.URL.Host+"/signup", 301)
				return
			}

			//insert into db
			if _, err := engine.Insert(user); err != nil {
				ctx.Redirect(ctx.Req.URL.Host+"/signup", 301)
				return
			}

			//now redirect to homepage
			ctx.Redirect(ctx.Req.URL.Host, 301)
		})

		m.Get("/signin", func(ctx *macaron.Context) {
			//if logged in, redirect to home
			if ctx.Data["Login"] == 1 {
				ctx.Redirect(ctx.Req.URL.Host, 301)
				return
			}
			ctx.HTML(200, "signin")
		})
		m.Post("/signin", binding.Bind(Users{}), func(ctx *macaron.Context, user Users, sess session.Store) {
			//if logged in, redirect to home
			if ctx.Data["Login"] == 1 {
				ctx.Redirect(ctx.Req.URL.Host, 301)
				return
			}
			//checks in DB
			//TODO: use encrypton
			//TODO: use super cookie secret

			//check if username exists
			//then check if password matches
			//then setcookie

			//check if username exists
			var tmpUser = Users{Username: user.Username}
			has, err := engine.Get(&tmpUser)

			if err != nil || has == false {
				ctx.Redirect(ctx.Req.URL.Host+"/signin", 301)
				return
			}

			//setcookie
			sess.Set("user", user.Username)

			//it is used to render corresponding html
			//if login set to 1 display logout option
			//else display signin signup option
			//		ctx.Data["Login"] = 1

			//redirect to home
			ctx.Redirect(ctx.Req.URL.Host, 301)
		})

		m.Get("/signout", func(ctx *macaron.Context, sess session.Store) {
			//if logged out, redirect to home
			fmt.Println("hello world")

			if ctx.Data["Login"] == 0 {
				ctx.Redirect(ctx.Req.URL.Host, 301)
				return
			}
			//remove the cookie
			//ctx.ClearCookie()

			//show signup sign option in html
			//ctx.Data["Login"] = 0
			/*
				cookie := &http.Cookie{
					Name:  "user",
					Value: "tahsin",
					//Expires: time.Now(),
				}
				http.SetCookie(ctx.Resp, cookie)
			*/
			fmt.Println("asfddddddddddddddddddddddddddd")
			sess.Delete("user")

			ctx.Redirect(ctx.Req.URL.Host, 301)
		})
	})

	m.Run()
}
