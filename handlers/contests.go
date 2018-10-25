package handlers

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
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

//route: /contests
//show all contests, sorted by time TODO: sorted by time
func GetAllContests(ctx *macaron.Context) {
	var all, running, upcoming, past []Contest

	if err := db.Engine.Find(&all); err != nil {
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	for _, contest := range all {
		contestStartTime, contestEndTime := contest.ContestStartTime, contest.ContestEndTime

		//catagorize into running, upcoming and past
		if time.Now().After(contestStartTime) {
			if contestEndTime.After(time.Now()) {
				running = append(running, contest)
			} else {
				past = append(past, contest)
			}
		} else {
			upcoming = append(upcoming, contest)
		}
	}

	//TODO: sort by time
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

func newContestForm(ctx *macaron.Context, contest Contest, update bool) (Contest, error) {
	//check if manager is valid
	var manager = Users{Username: contest.Manager}
	has, err := db.Engine.Get(&manager)

	if err != nil {
		//ctx.Resp.Write([]byte("500 internal server error"))
		return contest, err
	}

	if has == false {
		return contest, errors.New("manager not found")
	}

	//use namanger name instead of handle
	contest.Manager = manager.Username
	contest.ManagerId = manager.Id

	//update start and end time
	st, en, err := findTime(contest)

	if err != nil {
		//fmt.Println(err)
		return contest, err
	}

	//set data according to date format in db
	contest.ContestStartTime = st
	contest.ContestEndTime = en

	log.Println("============", contest)

	log.Println(time.Now())
	log.Println(st)
	log.Println(en)
	if !update && time.Now().After(st) {
		return contest, errors.New("start time must not be less than current time")
	}

	return contest, nil
}

//route: /contests/new POST
//create a new contest, admin must be logged in
func PostContest(ctx *macaron.Context, contest Contest) {
	newContest, err := newContestForm(ctx, contest, false)

	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	log.Println("===================", newContest)

	//insert the contest into the db
	_, err = db.Engine.Insert(&newContest)

	if err != nil {
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte("error here " + err.Error()))
		return
	}

	//redirect to contests page
	ctx.Redirect("/contests")

}

//route: /contests/:cid
//show dashboard/list of all problems if (contest is running && user has permission) || user = manager
//if logged in, show solved in green
func GetDashboard(ctx *macaron.Context) {
	//checkigs are done in middleware

	contest := ctx.Data["Contest"].(Contest)

	var all []Problem
	if err := db.Engine.Where("contest_id = ?", ctx.Params(":cid")).Find(&all); err != nil {
		//fmt.Println(err)
		//ctx.Resp.Write([]byte("500 internal server error"))
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	//TODO: optimize timing in dashboard

	//user == manager
	//user has permission && contest is running
	if ctx.Data["Username"] != nil && ctx.Data["Username"].(string) == contest.Manager {
		ctx.Data["Problems"] = all
	} else if time.Now().After(contest.ContestStartTime) {
		//check user has permission
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
func PutContest(ctx *macaron.Context, contest Contest) {
	newContest, err := newContestForm(ctx, contest, true)

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

	ctx.Redirect("/contests/" + cid)
}

//route: /contests/:cid DELETE
//show dashboard
//if logged in, show solved in green
func DeleteContest(ctx *macaron.Context) {
	fmt.Println("GetDashboard")
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
func GetAllSubmissions(ctx *macaron.Context) {
	fmt.Println("GetAllSubmissions")
}

//route: /contests/:cid/mysubmissions
//show my submission, if logged in and eligible
func GetMySubmissions(ctx *macaron.Context) {
	fmt.Println("GetMySubmissions")
}

func ContestAuth(ctx *macaron.Context) {
	username := ctx.Data["Username"].(string)
	contest := ctx.Data["Contest"].(Contest)
	db.Engine.Insert(&ContestPermission{UserName: username, ContestId: contest.Id})
}
