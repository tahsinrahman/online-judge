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
	ProblemId    int
	ContestId    int
	OrderId      string
	Name         string
	Description  string
	Input        string
	Output       string
	SampleInput  string
	SampleOutput string
	Notes        string
}

type Contest struct {
	ContestId        int
	Name             string    `form:"name"`
	StartDate        string    `form:"date"` //used to get data from form
	StartTime        string    `form:"time"` //used to get data from form
	Duration         string    `form:"duration"`
	ContestStartTime time.Time //saved in mysql after processing from startdate
	ContestEndTime   time.Time //saved in mysql after processing formdata
	Manager          string    `form:"manager"`
	ManagerId        int
}

func findTime(contest Contest) (time.Time, time.Time, error) {
	startDate := strings.Split(contest.StartDate, "-")
	startTime := strings.Split(contest.StartTime, ":")

	year, err := strconv.Atoi(startDate[0])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	month, err := strconv.Atoi(startDate[1])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	day, err := strconv.Atoi(startDate[2])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	hour, err := strconv.Atoi(startTime[0])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	min, err := strconv.Atoi(startTime[1])
	if err != nil {
		return time.Now(), time.Now(), err
	}

	//start time of contest
	contestStartTime := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local)

	duration := strings.Split(contest.Duration, ":")
	hour, err = strconv.Atoi(duration[0])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	min, err = strconv.Atoi(duration[1])
	if err != nil {
		return time.Now(), time.Now(), err
	}

	contestDuration := time.Duration(hour*60*60+min*60) * time.Second
	contestEndTime := contestStartTime.Add(contestDuration)

	return contestStartTime, contestEndTime, nil
}

//route: /contests
//show all contests, sorted by time TODO: sorted by time
//first show current contests, then future, then past
func GetAllContests(ctx *macaron.Context) {
	//show all contests from mysql to html
	var all, running, upcoming, past []Contest

	if err := db.Engine.Find(&all); err != nil {
		fmt.Println(err)
		ctx.Resp.Write([]byte("500 internal server error"))
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
	//only admin can view this page
	if ctx.Data["Username"] != "admin" {
		ctx.Resp.Write([]byte("unauthorized. only admin can create a new contest"))
		return
	}

	//if authorised then only can view the page
	ctx.HTML(200, "new_contest")
}

//route: /contests/new POST
//create a new contest, admin must be logged in
func PostNewContest(ctx *macaron.Context, contest Contest) {
	//only admin has this privilage
	if ctx.Data["Username"] != "admin" {
		ctx.Resp.Write([]byte("unauthorized. only admin can create a new contest"))
		return
	}

	//check if manager is valid
	var manager = Users{Username: contest.Manager}
	has, err := db.Engine.Get(&manager)

	if err != nil {
		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}

	if has == false {
		ctx.Resp.Write([]byte("manager not found"))
		return
	}

	//use namanger name instead of handle
	contest.Manager = manager.Name
	contest.ManagerId = manager.UserId

	//update start and end time
	st, en, err := findTime(contest)

	if err != nil {
		fmt.Println(err)
		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}

	//set data according to date format in db
	contest.ContestStartTime = st
	contest.ContestEndTime = en

	//insert the contest into the db
	_, err = db.Engine.Insert(&contest)

	if err != nil {
		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}

	//redirect to contests page
	ctx.Redirect("/contests")
}

//route: /contests/:cid
//show dashboard
//if logged in, show solved in green
func GetDashboard(ctx *macaron.Context) {
	//show all problems of this contest
	var all []Problem

	cid, err := strconv.Atoi(ctx.Params(":cid"))

	if err != nil {
		fmt.Println(err)
		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}

	if err := db.Engine.Where("contest_id = ?", cid).Find(&all); err != nil {
		fmt.Println(err)
		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}

	var contest = Contest{ContestId: cid}
	has, err := db.Engine.Get(&contest)

	if err != nil {
		fmt.Println(err)
		ctx.Resp.Write([]byte("500 internal server error"))
		return
	}

	if has == false {
		ctx.Resp.Write([]byte("contest doesn't exist"))
		return
	}

	//	contestStartTime, contestEndTime, err := findTime(contest)

	ctx.Data["Contest"] = contest
	ctx.Data["Problems"] = all
	ctx.Data["CurrentTime"] = time.Now().Format(time.RFC3339)

	//problems
	ctx.HTML(200, "dashboard")
}

//route: /contests/:cid/update GET
//update contest
//if logged in
func GetUpdateContest(ctx *macaron.Context) {
}

//route: /contests/:cid/update POST
//show contest
//if logged in
func PostUpdateContest(ctx *macaron.Context) {
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
