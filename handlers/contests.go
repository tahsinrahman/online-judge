package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

type JudgeData struct {
	id     int
	input  []byte
	output []byte
}

type Problem struct {
	id            int
	name          string
	description   []byte
	sample_input  []byte
	sample_output []byte
	notes         []byte
	JudgeData     []JudgeData
}

type Contest struct {
	Id        int
	Name      string `form:"name"`
	StartDate string `form:"date"`
	StartTime string `form:"time"`
	Duration  string `form:"duration"`
	Manager   string `form:"manager"`
	//problems []Problem
}

//route: /contests
//show all contests, sorted by time TODO: sorted by time
//first show current contests, then future, then past
func GetAllContests(ctx *macaron.Context) {
	//show all contests from mysql to html
	var all, running, upcoming, past []Contest

	if err := db.Engine.Find(&all); err != nil {
		//TODO: response internal server error
		ctx.Redirect("/")
		return
	}

	for _, contest := range all {
		startDate := strings.Split(contest.StartDate, "-")
		startTime := strings.Split(contest.StartTime, ":")

		year, err := strconv.Atoi(startDate[0])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}
		month, err := strconv.Atoi(startDate[1])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}
		day, err := strconv.Atoi(startDate[2])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}
		hour, err := strconv.Atoi(startTime[0])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}
		min, err := strconv.Atoi(startTime[1])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}
		sec, err := strconv.Atoi(startTime[2])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}

		//start time of contest
		contestStartTime := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)

		duration := strings.Split(contest.Duration, ":")
		hour, err = strconv.Atoi(duration[0])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}
		min, err = strconv.Atoi(duration[1])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}
		sec, err = strconv.Atoi(duration[2])
		if err != nil {
			//TODO: 500 error
			ctx.Redirect("/")
			return
		}

		contestDuration := time.Duration(hour*60*60+min*60+sec) * time.Second
		contestEndTime := contestStartTime.Add(contestDuration)

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
//TODO: only admin can create contest
//TODO: add option for participient list
func GetNewContest(ctx *macaron.Context) {
	ctx.HTML(200, "new_contest")
}

//route: /contests/new POST
//create a new contest, admin must be logged in
func PostNewContest(ctx *macaron.Context, contest Contest) {
	//insert the contest into the db
	_, err := db.Engine.Insert(&contest)

	if err != nil {
		//TODO: response internal server error
		ctx.Redirect("/contests/new")
		return
	}

	//redirect to contests page
	ctx.Redirect("/contests")
}

//route: /contests/:cid
//show dashboard
//if logged in, show solved in green
func GetContest(ctx *macaron.Context) {
	fmt.Println("GetDashboard")
}

//route: /contests/:cid PUT
//show dashboard
//if logged in, show solved in green
func UpdateContest(ctx *macaron.Context) {
	fmt.Println("GetDashboard")
}

//route: /contests/:cid DELETE
//show dashboard
//if logged in, show solved in green
func DeleteContest(ctx *macaron.Context) {
	fmt.Println("GetDashboard")
}

//route: /contests/:cid/:pid
//show problem
//show submit button if eligible and logged in
func GetProblem(ctx *macaron.Context) {
	fmt.Println("GetProblem")
}

//route: /contests/:cid/:pid PUT
func UpdateProblem(ctx *macaron.Context) {
	fmt.Println("GetProblem")
}

//route: /contests/:cid/:pid DELETE
func DeleteProblem(ctx *macaron.Context) {
	fmt.Println("GetProblem")
}

//route: /contests/:cid/:pid POST
//submit problem if eligible and logged in
func PostSubmit(ctx *macaron.Context) {
	fmt.Println("Post Submit")
}

func SubmitProblem(ctx *macaron.Context) {
	fmt.Println("Post Submit")
}

//route: /contests/:cid/rank
//show contest ranklist
func GetRank(ctx *macaron.Context) {
	fmt.Println("GetRank")
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
