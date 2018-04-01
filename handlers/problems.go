package handlers

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strconv"

	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

//Dataset has a set of input and output files
//each set has associated point values TODO
type Dataset struct {
	JudgeInput  []*multipart.FileHeader `form:"input_data[]"`
	JudgeOutput []*multipart.FileHeader `form:"output_data[]"`
	//Point       []int                   `form:"point[]"`
}

//structure of each problem
//judge data is seperated? Should be connected?
type Problem struct {
	ProblemId    int
	ContestId    int
	Name         string  `form:"name"`
	Description  string  `form:"description"`
	Input        string  `form:"input"`
	Output       string  `form:"output"`
	SampleInput  string  `form:"sample_input"`
	SampleOutput string  `form:"sample_output"`
	TimeLimit    float64 `form:"timelimit"`
	Notes        string  `form:"notes"`
	//OrderId      string
}

//route: /contests/:cid/new GET
//shows the html for for a new problem
func GetNewProblem(ctx *macaron.Context) {
	ctx.HTML(200, "new_problem")
}

//gets judge data as a file
//save them into local storage
//`dataset/cid/pid/in_id`
//`dataset/cid/pid/out_id`
func createFile(cid string, pid string, id string, in string, file *multipart.FileHeader) error {
	path := "dataset/" + cid + "/" + pid + "/" + in + "_" + id
	newInputFile, err := os.Create(path)

	if err != nil {
		return err
	}
	defer newInputFile.Close()

	f, err := file.Open()
	if err != nil {
		return nil
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil
	}

	_, err = newInputFile.Write(b)
	if err != nil {
		return err
	}
	return nil
}

//route: /contests/:cid/new POST
//gets problem info as formdata
//inserts infos into db and save files in system storage
//finally redirects to `contest/cid`
func PostNewProblem(ctx *macaron.Context, problem Problem, dataset Dataset) {
	contest := ctx.Data["Contest"].(Contest)

	problem.ContestId = contest.ContestId
	problem.ProblemId = contest.ProblemCount + 1
	contest.ProblemCount++

	//update contest in db
	db.Engine.Id(contest.ContestId).Update(&contest)

	//insert problem into db
	db.Engine.Insert(&problem)

	cid := strconv.Itoa(contest.ContestId)
	pid := strconv.Itoa(problem.ProblemId)

	//create a new direcory
	os.MkdirAll("dataset/"+cid+"/"+pid, 0755)

	//save each input/output file in system storage
	for i := 0; i < len(dataset.JudgeInput); i++ {
		if err := createFile(cid, pid, strconv.Itoa(i), "in", dataset.JudgeInput[i]); err != nil {
			ctx.Resp.Write([]byte(err.Error()))
			return
		}
		if err := createFile(cid, pid, strconv.Itoa(i), "out", dataset.JudgeOutput[i]); err != nil {
			ctx.Resp.Write([]byte(err.Error()))
			return
		}
	}

	//finally redirecto to `contests/cid`
	ctx.Redirect("/contests/" + cid)
}

func GetUpload(ctx *macaron.Context) {
	ctx.HTML(200, "test")
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
