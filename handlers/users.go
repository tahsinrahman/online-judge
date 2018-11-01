package handlers

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

type Users struct {
	Id        int64
	Name      string `form:"name"`
	Email     string `form:"email"`
	Username  string `form:"username"`
	Password  string `form:"password"`
	Privilege string `form:"privilege"`
	Approved  int
}

func Init() {
	db.Engine.Sync(new(Users))
	db.Engine.Sync(new(Problem))
	db.Engine.Sync(new(Contest))
	db.Engine.Sync(new(Dataset))
	db.Engine.Sync(new(Submission))
	db.Engine.Sync(new(Rank))
	db.Engine.Sync(new(ContestPermission))
	db.Engine.Sync(new(DatasetVerdict))
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

	if user.Privilege == "teacher" {
		user.Approved = 0
	} else {
		user.Approved = 1
	}

	//insert into db
	if _, err := db.Engine.Insert(user); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	//now redirect to homepage
	ctx.Redirect(ctx.Req.URL.Host, 302)
}

func WaitingForApproval(ctx *macaron.Context) {
	username := ctx.Data["Username"].(string)

	if username != "admin" {
		ctx.Resp.Write([]byte("unauthorized"))
	}

	var users []Users
	db.Engine.Where("approved = 0").Find(&users)

	ctx.Data["List"] = users

	ctx.HTML(200, "waiting_list")
}

func ApproveTeacher(ctx *macaron.Context) {
	username := ctx.Data["Username"].(string)

	if username != "admin" {
		ctx.Resp.Write([]byte("unauthorized"))
	}

	spew.Dump(ctx.Params("user"))

	var user Users
	db.Engine.Where("username = ?", ctx.Params("user")).Get(&user)

	user.Approved = 1
	db.Engine.Id(user.Id).Update(&user)

	ctx.Redirect("/wait")
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

	if tmpUser.Approved == 0 {
		ctx.Resp.Write([]byte("waiting for approval"))
		return
	}

	//setcookie
	ctx.SetSecureCookie("user", user.Username)

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
