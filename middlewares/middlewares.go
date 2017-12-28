package middlewares

import (
	"strconv"

	"github.com/tahsinrahman/online-judge/db"
	"github.com/tahsinrahman/online-judge/handlers"
	macaron "gopkg.in/macaron.v1"
)

//get username from cookie, check if username exists
//TODO: securecookie using token
func CheckAuthentication(ctx *macaron.Context) {
	cookie := ctx.GetCookie("user")

	if cookie == "" {
		return
	}

	//user if logged in
	//used it for showing logout option in html
	ctx.Data["Login"] = 1
	ctx.Data["Username"] = cookie
}

func CheckContestExistance(ctx *macaron.Context) {
	cid, err := strconv.Atoi(ctx.Params(":cid"))
	//	fmt.Println(cid, err)
	if err != nil {
		//		fmt.Println(err)
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	var contest = handlers.Contest{ContestId: cid}
	has, err := db.Engine.Get(&contest)

	if err != nil {
		//fmt.Println(err)
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	if has == false {
		ctx.Resp.Write([]byte("contest doesn't exist"))
		return
	}

	ctx.Data["Contest"] = contest
	//	fmt.Println("end of CheckContestExistance")
	//fmt.Println(contest)
}

func CheckManager(ctx *macaron.Context) {
	//	fmt.Println("start of checkmanager")
	contest, _ := ctx.Data["Contest"].(handlers.Contest)
	//	fmt.Println(contest, ctx.Data["Username"])

	if ctx.Data["Username"].(string) != contest.ManagerUsername {
		ctx.Resp.Write([]byte("unauthorized. only contest manager can update contest"))
		return
	}
}
