package handlers

import (
	"fmt"

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
//show all contests, sorted by time
//first show current contests, then future, then past
func GetAllContests(ctx *macaron.Context) {
	//show all contests from mysql to html
	var all []Contest

	if err := db.Engine.Find(&all); err != nil {
		//TODO: response internal server error
		fmt.Println("error")
		ctx.Redirect("/")
		return
	}

	ctx.Data["All"] = all

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
	fmt.Println(contest)
	_, err := db.Engine.Insert(&contest)

	if err != nil {
		//TODO: response internal server error
		fmt.Println(err)
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
