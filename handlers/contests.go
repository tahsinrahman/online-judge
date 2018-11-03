package handlers

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-macaron/cache"

	"github.com/davecgh/go-spew/spew"
	redisCache "github.com/tahsinrahman/online-judge/cache"
	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

//Contest format
type Contest struct {
	Id               int64
	Name             string    `form:"name"`
	StartDate        string    `form:"date"` //used to get data from form
	StartTime        string    `form:"time"` //used to get data from form
	Duration         string    `form:"duration"`
	ContestStartTime time.Time //saved in mysql after processing from startdate
	ContestEndTime   time.Time //saved in mysql after processing formdata
	Password         string    `form:"password"`
	Manager          string    `form:"manager"`
	ManagerId        int64
	ProblemCount     int
}

type ContestPermission struct {
	Id        int64
	UserName  string
	ContestId int64
}

//converts to time.Time from string
func findTime(contest Contest) (time.Time, time.Time, error) {
	startDate := strings.Split(contest.StartDate, "-")
	startTime := strings.Split(contest.StartTime, ":")

	tmp := time.Now()
	if len(startDate) != 3 {
		return tmp, tmp, errors.New("invalid date")
	}

	if len(startTime) != 2 {
		return tmp, tmp, errors.New("invalid time")
	}

	year, err := strconv.Atoi(startDate[0])
	if err != nil {
		return tmp, tmp, err
	}
	month, err := strconv.Atoi(startDate[1])
	if err != nil {
		return tmp, tmp, err
	}
	day, err := strconv.Atoi(startDate[2])
	if err != nil {
		return tmp, tmp, err
	}
	hour, err := strconv.Atoi(startTime[0])
	if err != nil {
		return tmp, tmp, err
	}
	min, err := strconv.Atoi(startTime[1])
	if err != nil {
		return tmp, tmp, err
	}

	//start time of contest
	contestStartTime := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local)

	duration := strings.Split(contest.Duration, ":")

	if len(duration) != 2 {
		return tmp, tmp, errors.New("invalid duration")
	}

	hour, err = strconv.Atoi(duration[0])
	if err != nil {
		return tmp, tmp, err
	}
	min, err = strconv.Atoi(duration[1])
	if err != nil {
		return tmp, tmp, err
	}

	contestDuration := time.Duration(hour*60*60+min*60) * time.Second
	contestEndTime := contestStartTime.Add(contestDuration)

	return contestStartTime, contestEndTime, nil
}
func CatagorizeContest(contest *Contest, running, upcoming, past *[]Contest) {
	contestStartTime, contestEndTime := contest.ContestStartTime, contest.ContestEndTime

	//catagorize into running, upcoming and past
	if time.Now().After(contestStartTime) {
		if contestEndTime.After(time.Now()) {
			*running = append(*running, *contest)
		} else {
			*past = append(*past, *contest)
		}
	} else {
		*upcoming = append(*upcoming, *contest)
	}
}

//route: /contests
//show all contests, sorted by time TODO: sorted by time
func GetAllContests(ctx *macaron.Context, c cache.Cache) {
	var all, running, upcoming, past []Contest

	key := "contest_list"
	if _, err := redisCache.FindObject(c, key, &all, db.Engine.NewSession(), false); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	for _, contest := range all {
		CatagorizeContest(&contest, &running, &upcoming, &past)
	}

	ctx.Data["Upcoming"] = upcoming
	ctx.Data["Running"] = running
	ctx.Data["Past"] = past

	//for rendering html
	ctx.HTML(200, "contests")
}

//route: /contests/new
//create a new contest, admin must be logged in
func GetNewContest(ctx *macaron.Context) {
	//if authorised then only can view the page
	ctx.HTML(200, "new_contest")
}

func newContestForm(ctx *macaron.Context, c cache.Cache, contest Contest, update bool) (Contest, error) {
	//check if manager is valid
	var manager = Users{Username: contest.Manager}

	key := fmt.Sprintf("user_%v", manager.Username)
	has, err := redisCache.FindObject(c, key, &manager, nil, true)

	if err != nil {
		return contest, err
	} else if has == false {
		return contest, errors.New("manager not found")
	}

	//use namanger name instead of handle
	contest.Manager = manager.Username
	contest.ManagerId = manager.Id

	//update start and end time
	st, en, err := findTime(contest)

	if err != nil {
		return contest, err
	}

	//set data according to date format in db
	contest.ContestStartTime = st
	contest.ContestEndTime = en

	if !update && time.Now().After(st) {
		return contest, errors.New("start time must not be less than current time")
	}

	return contest, nil
}

//route: /contests/new POST
//create a new contest, admin must be logged in
func PostContest(ctx *macaron.Context, c cache.Cache, contest Contest) {
	newContest, err := newContestForm(ctx, c, contest, false)

	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// insert the contest into the db
	_, err = db.Engine.Insert(&newContest)

	if err != nil {
		ctx.Resp.Write([]byte("error here " + err.Error()))
		return
	}

	// remove old cache with key "contest_list"
	key := "contest_list"
	if err = c.Delete(key); err != nil {
		ctx.Resp.Write([]byte("error here " + err.Error()))
		return
	}

	//redirect to contests page
	ctx.Redirect("/contests")
}

//route: /contests/:cid
//show dashboard/list of all problems if (contest is running && user has permission) || user = manager
//if logged in, show solved in green
func GetDashboard(ctx *macaron.Context, c cache.Cache) {
	//checkigs are done in middleware
	contest := ctx.Data["Contest"].(Contest)

	key := fmt.Sprintf("contest_%v_problems", contest.Id)
	dbSession := db.Engine.Where("contest_id = ?", ctx.Params(":cid"))
	var all []Problem

	if _, err := redisCache.FindObject(c, key, &all, dbSession, false); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// user == manager
	// user has permission && contest is running
	if ctx.Data["Username"] != nil && ctx.Data["Username"].(string) == contest.Manager {
		ctx.Data["Problems"] = all
	} else if time.Now().After(contest.ContestStartTime) {
		// check user has permission
		ctx.Data["Problems"] = all
	}

	//problems
	ctx.HTML(200, "dashboard")
}

//route: /contests/:cid/update GET
//update contest
func GetUpdateContest(ctx *macaron.Context) {
	ctx.HTML(200, "update_contest")
}

//route: /contests/:cid/update POST
//update contest info into db
func PutContest(ctx *macaron.Context, c cache.Cache, contest Contest) {
	newContest, err := newContestForm(ctx, c, contest, true)

	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	cid := ctx.Params(":cid")

	//update db
	_, err = db.Engine.Id(cid).Update(&newContest)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// remove old contest from cache with key "contest=ID"
	key := fmt.Sprintf("contest_%v", cid)
	if err = c.Delete(key); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	// remove old cache with key "contest_list"
	if err = c.Delete("contest_list"); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	ctx.Redirect("/contests/" + cid)
}

//route: /contests/:cid/rank
//show contest ranklist
func GetRank(ctx *macaron.Context) {
	cid := ctx.Params(":cid")

	contestId, err := strconv.ParseInt(cid, 10, 64)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// find users who have participated
	var users []string
	err = db.Engine.Table("rank").Where("contest_id = ?", contestId).Cols("user_name").Find(&users)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	spew.Dump(users)

	unique := make(map[string]int)

	type UserRank struct {
		TotalScore    float64
		TotalPenalty  int64
		ProblemScores []Rank
	}
	var ranklist []UserRank

	// now, get list of problems solved by each user
	for _, username := range users {
		if unique[username] == 1 {
			continue
		}
		unique[username] = 1

		spew.Dump(username)

		var rank []Rank
		err = db.Engine.Find(&rank, &Rank{ContestId: contestId, UserName: username})
		if err != nil {
			ctx.Resp.Write([]byte(err.Error()))
			return
		}

		var totalScore float64
		var totalPenalty int64
		for _, myrank := range rank {
			totalScore += myrank.Score
			totalPenalty += myrank.Penalty
		}

		userRank := UserRank{
			TotalScore:    totalScore,
			TotalPenalty:  totalPenalty,
			ProblemScores: rank,
		}

		ranklist = append(ranklist, userRank)
	}
	spew.Dump(ranklist)

	sort.Slice(ranklist, func(i, j int) bool {
		if ranklist[i].TotalScore == ranklist[j].TotalScore {
			return ranklist[i].TotalPenalty < ranklist[j].TotalPenalty
		}
		return ranklist[i].TotalScore > ranklist[j].TotalScore
	})
	spew.Dump(ranklist)
	ctx.Data["ranklist"] = ranklist
	ctx.HTML(200, "ranklist")
}

//route: /contests/:cid/allsubmissions
//show all submissions, sorted by time
//maybe add filter option
type ContestSubmission struct {
	Submission `xorm:"extends"`
	Name       string
	ContestId  int64
}

func (ContestSubmission) TableName() string {
	return "submission"
}

func GetAllSubmissions(ctx *macaron.Context) {
	key := fmt.Sprintf("contest_%v_submissions", ctx.Params(":cid"))
	dbSession := db.Engine.Join("INNER", "problem", "problem.id = submission.problem_id").Where("problem.contest_id = ?", ctx.Params("cid"))
	var submissions []ContestSubmission

	if err := redisCache.CheckList(key, &submissions, dbSession); err != nil {
		ctx.Resp.Write([]byte("please log-in"))
		return
	}

	ctx.Data["Submissions"] = submissions

	ctx.HTML(200, "all_submissions")
}

func ContestAuth(ctx *macaron.Context) {
	username, _ := ctx.Data["Username"].(string)

	if username == "" {
		ctx.Resp.Write([]byte("please log-in"))
		return
	}

	contest := ctx.Data["Contest"].(Contest)
	db.Engine.Insert(&ContestPermission{UserName: username, ContestId: contest.Id})
}
