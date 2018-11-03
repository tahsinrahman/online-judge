package middlewares

import (
	"fmt"
	"strconv"
	"time"

	redisCache "github.com/tahsinrahman/online-judge/cache"

	"github.com/go-macaron/cache"

	"github.com/tahsinrahman/online-judge/db"
	"github.com/tahsinrahman/online-judge/handlers"
	macaron "gopkg.in/macaron.v1"
)

// get username from cookie, check if username exists in database
// if username not found, user not authenticated
func CheckAuthentication(ctx *macaron.Context, c cache.Cache) {
	cookie, has := ctx.GetSecureCookie("user")

	if !has {
		return
	}

	user := handlers.Users{Username: cookie}
	key := fmt.Sprintf("user_%v", user.Username)

	has, err := redisCache.FindObject(c, key, &user, nil, true)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// found user
	if has {
		ctx.Data["Login"] = 1
		ctx.Data["Username"] = cookie
		ctx.Data["Previlege"] = user.Privilege
	}
}

// check if contest id exists
func CheckContestExistance(ctx *macaron.Context, c cache.Cache) {
	cid, err := strconv.Atoi(ctx.Params(":cid"))
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	contest := handlers.Contest{Id: int64(cid)}
	key := fmt.Sprintf("contest_%v", contest.Id)

	has, err := redisCache.FindObject(c, key, &contest, nil, true)

	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	if !has {
		ctx.Resp.Write([]byte("contest doesn't exist"))
		return
	}

	ctx.Data["Contest"] = contest
}

func CheckProblem(ctx *macaron.Context, c cache.Cache) {
	pid, err := strconv.Atoi(ctx.Params(":pid"))
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	contest := ctx.Data["Contest"].(handlers.Contest)

	problem := handlers.Problem{ContestId: contest.Id, ProblemId: pid}
	key := fmt.Sprintf("contest_%v_problem_%v", problem.ContestId, problem.ProblemId)

	has, err := redisCache.FindObject(c, key, &problem, nil, true)

	if err != nil {
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

func AddTests(ctx *macaron.Context, c cache.Cache) {
	problem, _ := ctx.Data["Problem"].(handlers.Problem)
	var dataset []handlers.Dataset

	key := fmt.Sprintf("problem_%v_dataset", problem.Id)
	xormSession := db.Engine.Cols("id", "label", "weight").Where("problem_id = ?", problem.Id)

	if _, err := redisCache.FindObject(c, key, &dataset, xormSession, false); err != nil {
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

	key := fmt.Sprintf("submissions_user_%v_problem_%v", username, problem.Id)
	dbSession := db.Engine.Where("problem_id = ? and user_name = ?", problem.Id, username)
	var submissions []handlers.Submission

	if err := redisCache.CheckList(key, &submissions, dbSession); err != nil {
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

func AddContestPermission(ctx *macaron.Context, c cache.Cache) {
	if ctx.Data["Username"] != nil {
		username := ctx.Data["Username"].(string)
		contest := ctx.Data["Contest"].(handlers.Contest)

		perm := handlers.ContestPermission{UserName: username, ContestId: contest.Id}
		key := fmt.Sprintf("perm_contest_%v_user=%v", contest.Id, username)

		has, err := redisCache.FindObject(c, key, &perm, nil, true)
		if err != nil {
			ctx.Resp.Write([]byte(err.Error()))
			return
		}

		ctx.Data["Permission"] = has
	}
}
