package handlers

import (
	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

type Users struct {
	Name     string `form:"name"`
	Username string `form:"username"`
	Password string `form:"password"`
}

//homepage
func GetHome(ctx *macaron.Context) {
	//show index.html
	ctx.HTML(200, "index")
}

//signup form
func GetSignUp(ctx *macaron.Context) {
	//if logged in, redirect to home
	if ctx.Data["Login"] == 1 {
		ctx.Redirect(ctx.Req.URL.Host, 302)
		return
	}

	//show signup form
	ctx.HTML(200, "signup")
}

//get signup form data
func PostSignUp(ctx *macaron.Context, user Users) {
	//if logged in, redirect to home
	if ctx.Data["Login"] == 1 {
		ctx.Redirect(ctx.Req.URL.Host, 302)
		return
	}
	//TODO: encryption for pass

	//check if username already exists
	//TODO: if yes, then response with "username exists"
	//else insert new user into db

	//check if exists
	var tmpUser = Users{Username: user.Username}
	has, err := db.Engine.Get(&tmpUser)

	if err != nil || has {
		ctx.Redirect(ctx.Req.URL.Host+"/signup", 302)
		return
	}

	//insert into db
	if _, err := db.Engine.Insert(user); err != nil {
		//TODO: response internal server error
		ctx.Redirect(ctx.Req.URL.Host+"/signup", 302)
		return
	}

	//now redirect to homepage
	ctx.Redirect(ctx.Req.URL.Host, 302)
}

//signin form
func GetSignIn(ctx *macaron.Context) {
	//if logged in, redirect to home
	if ctx.Data["Login"] == 1 {
		ctx.Redirect(ctx.Req.URL.Host, 302)
		return
	}

	//show signin form
	ctx.HTML(200, "signin")
}

//get user data from form
func PostSignIn(ctx *macaron.Context, user Users) {
	//if logged in, redirect to home
	if ctx.Data["Login"] == 1 {
		ctx.Redirect(ctx.Req.URL.Host, 302)
		return
	}

	//checks in DB
	//TODO: use encrypton

	//check if username exists
	//then check if password matches
	//then setcookie

	//TODO: if username exists, show message "username exists", if password mismatch, show "wrong password"

	//check if username exists
	var tmpUser = Users{Username: user.Username, Password: user.Password}
	has, err := db.Engine.Get(&tmpUser)

	if err != nil || has == false {
		ctx.Redirect(ctx.Req.URL.Host+"/signin", 302)
		return
	}

	//setcookie
	//TODO: set secure cookie
	ctx.SetCookie("user", user.Username)

	//redirect to home
	ctx.Redirect(ctx.Req.URL.Host, 302)
}

//logout current user
func GetSignOut(ctx *macaron.Context) {
	//if logged out already, then redirect to home
	if ctx.Data["Login"] == 0 {
		ctx.Redirect(ctx.Req.URL.Host, 302)
		return
	}

	//deletes the cookie and redirects to home
	ctx.SetCookie("user", "")
	ctx.Redirect(ctx.Req.URL.Host, 302)
}
