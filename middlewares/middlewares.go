package middlewares

import (
	"strconv"
	"time"

	"github.com/tahsinrahman/online-judge/db"
	"github.com/tahsinrahman/online-judge/handlers"
	macaron "gopkg.in/macaron.v1"
)

//get username from cookie, check if username exists
//TODO: securecookie using token
func CheckAuthentication(ctx *macaron.Context) {
	cookie, has := ctx.GetSecureCookie("user")

	if !has {
		return
	}

	//user if logged in
	//used it for showing logout option in html
	ctx.Data["Login"] = 1
	ctx.Data["Username"] = cookie
}

func CheckContestExistance(ctx *macaron.Context) {
	tmp, err := strconv.Atoi(ctx.Params(":cid"))
	if err != nil {
		//		fmt.Println(err)
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	cid := int64(tmp)

	var contest = handlers.Contest{Id: cid}
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
}

func CheckProblem(ctx *macaron.Context) {
	pid, err := strconv.Atoi(ctx.Params(":pid"))
	if err != nil {
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	contest := ctx.Data["Contest"].(handlers.Contest)

	var problem = handlers.Problem{ContestId: contest.Id, ProblemId: pid}
	has, err := db.Engine.Get(&problem)

	if err != nil {
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	if has == false {
		ctx.Resp.Write([]byte("problem doesn't exist"))
		return
	}

	ctx.Data["Problem"] = problem
}

func CheckManager(ctx *macaron.Context) {
	contest, _ := ctx.Data["Contest"].(handlers.Contest)

	if ctx.Data["Username"] != nil && ctx.Data["Username"].(string) != contest.Manager {
		ctx.Resp.Write([]byte("unauthorized. only contest manager can update contest"))
		return
	}
}

func AddTests(ctx *macaron.Context) {
	problem, _ := ctx.Data["Problem"].(handlers.Problem)

	var dataset []handlers.Dataset

	err := db.Engine.Find(&dataset, &handlers.Dataset{ProblemId: problem.Id})
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	ctx.Data["Dataset"] = dataset
}

func AddSubmissions(ctx *macaron.Context) {
	problem, _ := ctx.Data["Problem"].(handlers.Problem)
	username, _ := ctx.Data["Username"].(string)

	if username == "" {
		return
	}

	var submissions []handlers.Submission
	err := db.Engine.Find(&submissions, &handlers.Submission{ProblemId: problem.Id, UserName: username})

	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	ctx.Data["Submissions"] = submissions
}

func CheckStartTime(ctx *macaron.Context) {
	contest, _ := ctx.Data["Contest"].(handlers.Contest)

	if time.Now().After(contest.ContestStartTime) {
		return
	}

	if ctx.Data["Username"] != nil && ctx.Data["Username"].(string) != contest.Manager {
		ctx.Resp.Write([]byte("unauthorized. only contest manager can update contest"))
		return
	}
}

func CheckAdmin(ctx *macaron.Context) {
	//only admin has this privilage
	if ctx.Data["Username"] != "admin" {
		ctx.Resp.Write([]byte("unauthorized. only admin can create a new contest"))
		return
	}
}

func CheckEndTime(ctx *macaron.Context) {
	contest, _ := ctx.Data["Contest"].(handlers.Contest)

	if time.Now().Before(contest.ContestEndTime) {
		return
	}
	if ctx.Data["Username"] != nil && ctx.Data["Username"].(string) != contest.Manager {
		ctx.Resp.Write([]byte("contest ended"))
		return
	}
}

func AddContestPermission(ctx *macaron.Context) {
	if ctx.Data["Username"] != nil {
		username := ctx.Data["Username"].(string)
		contest := ctx.Data["Contest"].(handlers.Contest)

		has, err := db.Engine.Get(&handlers.ContestPermission{UserName: username, ContestId: contest.Id})

		if err != nil {
			ctx.Resp.Write([]byte(err.Error()))
			return
		}

		ctx.Data["Permission"] = has
	}
}
