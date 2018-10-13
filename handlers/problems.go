package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

//Dataset has a set of input and output files
//each set has associated point values TODO
type Dataset struct {
	Id          int64
	ProblemId   int
	Label       string
	JudgeInput  string `xorm:"text"`
	JudgeOutput string `xorm:"text"`
	Weight      int
}

type ProblemDataset struct {
	Label       []string                `form:"label[]"`
	JudgeInput  []*multipart.FileHeader `form:"input_data[]"`
	JudgeOutput []*multipart.FileHeader `form:"output_data[]"`
	Weight      []int                   `form:"weight[]"`
}

type Submission struct {
	Language string                `form:"language"`
	Source   *multipart.FileHeader `form:"source"`
}

//structure of each problem
//judge data is seperated? Should be connected?
type Problem struct {
	Id           int64
	ProblemId    int
	ContestId    int64
	Name         string  `form:"name"`
	Description  string  `form:"description" xorm:"varchar(1000)"`
	Input        string  `form:"input" xorm:"varchar(1000)"`
	Output       string  `form:"output" xorm:"varchar(1000)"`
	SampleInput  string  `form:"sample_input" xorm:"varchar(1000)"`
	SampleOutput string  `form:"sample_output" xorm:"varchar(1000)"`
	TimeLimit    float64 `form:"timelimit"`
	MemoryLimit  int     `form:"memorylimit"`
	Notes        string  `form:"notes" xorm:"varchar(1000)"`
	MaxPoint     int     `form:"maxpoint"`
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
func createFile(cid string, pid string, id string, in string, file *multipart.FileHeader) (string, error) {
	path := filepath.Join("dataset", cid, pid, id)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}

	path = filepath.Join(path, in)
	newInputFile, err := os.Create(path)

	if err != nil {
		return "", err
	}
	defer newInputFile.Close()

	f, err := file.Open()
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	_, err = newInputFile.Write(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func addDataset(cid, pid string, problemId int, dataset ProblemDataset) error {
	// add datasets
	for i, j := 0, len(dataset.JudgeInput); i < j; i++ {
		input, err := createFile(cid, pid, strconv.Itoa(i), "in", dataset.JudgeInput[i])
		if err != nil {
			return err
		}
		output, err := createFile(cid, pid, strconv.Itoa(i), "out", dataset.JudgeOutput[i])
		if err != nil {
			return err
		}

		data := Dataset{
			Label:       dataset.Label[i],
			Weight:      dataset.Weight[i],
			JudgeInput:  input,
			JudgeOutput: output,
			ProblemId:   problemId,
		}
		_, err = db.Engine.Insert(&data)
		if err != nil {
			return err
		}
	}
	return nil
}

//route: /contests/:cid/new POST
//gets problem info as formdata
//inserts infos into db and save files in system storage
//finally redirects to `contest/cid`
func PostProblem(ctx *macaron.Context, problem Problem, dataset ProblemDataset) {
	contest := ctx.Data["Contest"].(Contest)

	problem.ContestId = contest.Id
	problem.ProblemId = contest.ProblemCount + 1
	contest.ProblemCount++

	// insert problem into db
	_, err := db.Engine.Insert(&problem)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	cid := strconv.FormatInt(contest.Id, 10)
	pid := strconv.Itoa(problem.ProblemId)

	err = addDataset(cid, pid, problem.ProblemId, dataset)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// update contest in db
	_, err = db.Engine.Id(contest.Id).Update(&contest)
	if err != nil {
		// remove problem
		// remove datasets
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	//finally redirecto to dashboard
	ctx.Redirect("/contests/" + cid)
}

//route: /contests/:cid/:pid
//show problem
//show submit button if eligible and logged in
func GetProblem(ctx *macaron.Context) {
	ctx.HTML(200, "problem")
}

//route: /contests/:cid/:pid/update GET
func UpdateProblem(ctx *macaron.Context) {
	ctx.HTML(201, "update_problem")
}

//route: /contests/:cid/:pid/update PUT
func UpdateProblemDescripton(ctx *macaron.Context, problem Problem) {
	cid, err := strconv.Atoi(ctx.Params(":cid"))
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	pid, err := strconv.Atoi(ctx.Params(":pid"))
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	_, err = db.Engine.Update(&problem, &Problem{ProblemId: pid, ContestId: int64(cid)})

	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	ctx.Redirect("/contests/" + ctx.Params(":cid") + "/" + ctx.Params(":pid"))
}

func DownloadTest(ctx *macaron.Context) {
	id := ctx.Params(":id")
	dataType := ctx.Params(":type")

	var dataset Dataset

	found, err := db.Engine.Id(id).Get(&dataset)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	if !found {
		ctx.Resp.Write([]byte("not found"))
		return
	}

	//spew.Dump(dataset)

	var dload string
	if dataType == "in" {
		dload = dataset.JudgeInput
	} else {
		dload = dataset.JudgeOutput
	}

	ctx.Resp.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.txt", dataset.Label, dataType))

	io.Copy(ctx.Resp, strings.NewReader(dload))
	//ctx.Resp.Write([]byte(dload))
}

func GetList(ctx *macaron.Context) {
	dataset := ctx.Data["Dataset"].([]Dataset)
	err := json.NewEncoder(ctx.Resp).Encode(dataset)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
}

func DeleteTest(ctx *macaron.Context) {
	id := ctx.Params(":id")
	_, err := db.Engine.Id(id).Delete(&Dataset{})
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
}

func AddNewTest(ctx *macaron.Context, dataset ProblemDataset) {
	if len(dataset.JudgeInput) == 0 {
		ctx.Resp.Write([]byte("nothing found"))
		return
	}

	problem := ctx.Data["Problem"].(Problem)

	err := addDataset(ctx.Params(":cid"), ctx.Params(":pid"), problem.ProblemId, dataset)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
}

func UpdateProblemTests(ctx *macaron.Context) {
}
func UpdateProblemLimits(ctx *macaron.Context) {
}

//route: /contests/:cid/:pid DELETE
func DeleteProblem(ctx *macaron.Context) {
	fmt.Println("GetProblem")
}

//route: /contests/:cid/:pid POST
//submit problem if eligible and logged in
func SubmitProblem(solution Submission, ctx *macaron.Context) {
	runSubmission := func(solution Submission) {
		switch solution.Language {
		case "java":
		case "c++":
			//cmd := 'docker run --rm -v /home/tahsin/code/codeforces:/submission -w /submission debian /bin/sh -c "timeout 2 ./a.out"'
			//strings.Split(cmd)
		case "c":
		default:
			panic("unrecognized format")
		}
	}

	// save to local storage

	// acknowledge

	// run go routine
	go runSubmission(solution)

	cid := ctx.Params(":cid")
	pid := ctx.Params(":pid")

	ctx.Redirect("/contests/" + cid + "/" + pid)
}
