package main

import (
	"fmt"

	"github.com/go-macaron/binding"
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
func checkAuthenticationHandler(ctx *macaron.Context) {
	cookie := ctx.GetCookie("user")

	if cookie == "" {
		return
	}

	//user if logged in
	//used it for showing logout option in html
	ctx.Data["Login"] = 1
	ctx.Data["Username"] = cookie
}

func main() {
	//starts database engine
	initEngine()

	m := macaron.Classic()
	m.Use(macaron.Renderer()) //for html rendering
	m.Use(checkAuthenticationHandler)

	//homepage
	m.Get("/", func(ctx *macaron.Context) {
		ctx.HTML(200, "index")
	})

	//signup form
	m.Get("/signup", func(ctx *macaron.Context) {
		//if logged in, redirect to home
		if ctx.Data["Login"] == 1 {
			ctx.Redirect(ctx.Req.URL.Host, 302)
			return
		}

		ctx.HTML(200, "signup")
	})

	//create new user
	m.Post("/signup", binding.Bind(Users{}), func(ctx *macaron.Context, user Users) {
		//if logged in, redirect to home
		if ctx.Data["Login"] == 1 {
			ctx.Redirect(ctx.Req.URL.Host, 302)
			return
		}
		//insert to DB
		//TODO: encryption for pass

		//check if username already exists
		//TODO: if yes, then response with "username exists"
		//else insert new user into db

		//check if exists
		var tmpUser = Users{Username: user.Username}
		has, err := engine.Get(&tmpUser)

		if err != nil || has {
			ctx.Redirect(ctx.Req.URL.Host+"/signup", 302)
			return
		}

		//insert into db
		if _, err := engine.Insert(user); err != nil {
			ctx.Redirect(ctx.Req.URL.Host+"/signup", 302)
			return
		}

		//now redirect to homepage
		ctx.Redirect(ctx.Req.URL.Host, 302)
	})

	m.Get("/signin", func(ctx *macaron.Context) {
		//if logged in, redirect to home
		fmt.Println("hello signin")
		if ctx.Data["Login"] == 1 {
			ctx.Redirect(ctx.Req.URL.Host, 302)
			return
		}
		ctx.HTML(200, "signin")
	})

	m.Post("/signin", binding.Bind(Users{}), func(ctx *macaron.Context, user Users) {
		//if logged in, redirect to home
		if ctx.Data["Login"] == 1 {
			ctx.Redirect(ctx.Req.URL.Host, 302)
			return
		}
		//checks in DB
		//TODO: use encrypton
		//TODO: use super cookie secret

		//check if username exists
		//then check if password matches
		//then setcookie

		//TODO: if username exists, show message "username exists", if password mismatch, show "wrong password"

		//check if username exists
		var tmpUser = Users{Username: user.Username, Password: user.Password}
		has, err := engine.Get(&tmpUser)

		if err != nil || has == false {
			ctx.Redirect(ctx.Req.URL.Host+"/signin", 302)
			return
		}

		//setcookie
		ctx.SetCookie("user", user.Username)

		//redirect to home
		ctx.Redirect(ctx.Req.URL.Host, 302)
	})

	//TODO: logout, signout doesn't work, why?
	m.Get("/signout", func(ctx *macaron.Context) {
		//if logged out already, then redirect to home
		if ctx.Data["Login"] == 0 {
			ctx.Redirect(ctx.Req.URL.Host, 302)
			return
		}

		//deletes the cookie and redirects to home
		ctx.SetCookie("user", "")
		ctx.Redirect(ctx.Req.URL.Host, 302)
	})

	m.Run()
}
