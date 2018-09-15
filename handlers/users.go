package handlers

import (
	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

type Users struct {
	UserId   int    `xorm:"pk"`
	Name     string `form:"name"`
	Username string `form:"username"`
	Password string `form:"password"`
	Teacher  bool
}

func Init() {
	db.Engine.Sync(new(Users))
	db.Engine.Sync(new(Problem))
	db.Engine.Sync(new(Contest))
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
	//else insert new user into db

	//check if exists
	var tmpUser = Users{Username: user.Username}
	has, err := db.Engine.Get(&tmpUser)

	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		//		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}

	if has {
		ctx.Resp.Write([]byte("username exists"))
		return
	}

	//insert into db
	if _, err := db.Engine.Insert(user); err != nil {
		ctx.Resp.Write([]byte("500 internal server error"))
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

	if err != nil {
		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}
	if has == false {
		ctx.Resp.Write([]byte("username or password mismatch"))
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
