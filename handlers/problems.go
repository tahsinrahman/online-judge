package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

//Dataset has a set of input and output files
//each set has associated point values TODO
type Dataset struct {
	Id          int64
	ProblemId   int64
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
	Id         int64
	ProblemId  int64
	UserName   string
	Time       time.Time
	Language   string                `form:"language"`
	Source     *multipart.FileHeader `form:"source" xorm:"-"`
	Submission string                `xorm:"text"`
	Status     string
	Points     float64
}

type Rank struct {
	Id        int64
	ContestId int64
	ProblemId int64
	UserName  string
	Tries     int
	Penalty   int64
	Score     float64
	Status    int
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

func fileToByte(file *multipart.FileHeader) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return []byte{}, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

//gets judge data as a file
//save them into local storage
//`dataset/cid/pid/in_id`
//`dataset/cid/pid/out_id`

func createFile(path, filename string, file *multipart.FileHeader) (string, error) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}

	path = filepath.Join(path, filename)
	newInputFile, err := os.Create(path)

	if err != nil {
		return "", err
	}
	defer newInputFile.Close()

	b, err := fileToByte(file)
	if err != nil {
		return "", err
	}

	_, err = newInputFile.Write(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func addDataset(cid, pid string, problemId int64, dataset ProblemDataset) error {
	// add datasets
	for i, j := 0, len(dataset.JudgeInput); i < j; i++ {

		input, err := fileToByte(dataset.JudgeInput[i])
		if err != nil {
			return err
		}
		output, err := fileToByte(dataset.JudgeOutput[i])
		if err != nil {
			return err
		}

		data := Dataset{
			Label:       dataset.Label[i],
			Weight:      dataset.Weight[i],
			JudgeInput:  string(input),
			JudgeOutput: string(output),
			ProblemId:   problemId,
		}

		_, err = db.Engine.Insert(&data)
		if err != nil {
			return err
		}

		path := filepath.Join("dataset", cid, pid, strconv.FormatInt(data.Id, 10))
		_, err = createFile(path, "in", dataset.JudgeInput[i])
		if err != nil {
			return err
		}
		_, err = createFile(path, "out", dataset.JudgeOutput[i])
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

	err = addDataset(cid, pid, problem.Id, dataset)
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
	cid := ctx.Params(":cid")
	pid := ctx.Params(":pid")

	// remove from db
	_, err := db.Engine.Id(id).Delete(&Dataset{})
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	// remove from file system
	err = os.RemoveAll(filepath.Join("dataset", cid, pid, id))
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

	err := addDataset(ctx.Params(":cid"), ctx.Params(":pid"), problem.Id, dataset)
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

func updateSubmissionStatus(submission *Submission, contest *Contest, problem *Problem) error {
	// now
	rank := Rank{
		ContestId: contest.Id,
		ProblemId: problem.Id,
		UserName:  submission.UserName,
	}
	has, err := db.Engine.Get(&rank)
	if err != nil {
		return err
	}

	spew.Dump(rank)
	if has {
		if rank.Score < submission.Points {
			rank.Score = submission.Points
			rank.Penalty = int64(contest.ContestStartTime.Sub(submission.Time).Minutes()) + int64(rank.Tries*20)
			rank.Tries++
		} else {
			rank.Tries++
		}

		spew.Dump(int(rank.Score), problem.MaxPoint)
		if int(rank.Score) == problem.MaxPoint {
			rank.Status = 1
			spew.Dump(rank.Status)
		}

		// update
		_, err = db.Engine.Id(rank.Id).Update(&rank)
		if err != nil {
			return err
		}
	} else {
		rank.Score = submission.Points
		rank.Tries = 1
		if rank.Score > 0 {
			rank.Penalty = int64(contest.ContestStartTime.Sub(submission.Time).Minutes())
		}

		// insert
		_, err := db.Engine.InsertOne(&rank)
		if err != nil {
			panic(err)
		}
	}

	_, err = db.Engine.Id(submission.Id).Update(submission)
	return err
}

func compile(compileCmd string, submission *Submission, contest *Contest, problem *Problem) bool {
	args := strings.Split(compileCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Run()
	if err != nil {
		submission.Status = "compilation error"
		err = updateSubmissionStatus(submission, contest, problem)
		if err != nil {
			panic(err)
		}
		return false
	}
	return true
}

//route: /contests/:cid/:pid POST
//submit problem if eligible and logged in
func SubmitProblem(submission Submission, ctx *macaron.Context) {
	runSubmission := func(dataPath, sourcePath, filename, execName, timelimit string, solution Submission, problem *Problem, contest *Contest) {
		switch solution.Language {
		case "java":
		case "c++":
			compileCmd := fmt.Sprintf("g++ -O3 -std=c++14 -o %s %s", execName, filename)
			if !compile(compileCmd, &submission, contest, problem) {
				return
			}

			// now run for all tests
			var data []Dataset
			err := db.Engine.Find(&data, &Dataset{ProblemId: submission.ProblemId})
			if err != nil {
				panic(err)
			}

			total, maxPossible := 0, 0
			for _, dataset := range data {
				submissionId := strconv.FormatInt(submission.Id, 10)
				datasetId := strconv.FormatInt(dataset.Id, 10)

				cmdRun := exec.Command("/bin/bash", "scripts/runTest.sh", dataPath, submissionId, datasetId, timelimit)

				var buf bytes.Buffer
				cmdRun.Stdout = &buf

				err = cmdRun.Run()
				if err == nil {
					total += dataset.Weight
				}
				if buf.Len() != 0 && len(submission.Status) != 0 {
					submission.Status = buf.String()
					if submission.Status == "139\n" {
						submission.Status = "Runtime Error"
					} else if submission.Status == "WA\n" {
						submission.Status = "Wrong Answer"
					} else {
						submission.Status = "Time Limit Exceeded"
					}
				}
				maxPossible += dataset.Weight
			}

			if total == maxPossible {
				submission.Status = "Accepted"
			}

			submission.Points = float64(problem.MaxPoint) * float64(total) / float64(maxPossible)
			err = updateSubmissionStatus(&submission, contest, problem)
			if err != nil {
				panic(err)
			}

			// update database
		case "c":
			compileCmd := fmt.Sprintf("gcc -O3 -o %s %s", execName, filename)
			compile(compileCmd, &submission, contest, problem)
		default:
			panic("unrecognized format")
		}
	}

	username := ctx.Data["Username"].(string)
	problem := ctx.Data["Problem"].(Problem)
	contest := ctx.Data["Contest"].(Contest)

	// prepare submission
	submission.Time = time.Now()
	submission.UserName = username
	submission.ProblemId = problem.Id

	b, err := fileToByte(submission.Source)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	submission.Submission = string(b)
	submission.Status = "pending"

	// save to db
	_, err = db.Engine.Insert(&submission)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// save to local storage
	submissionId := strconv.FormatInt(submission.Id, 10)
	path := filepath.Join("submissions", submissionId)
	_, err = createFile(path, submission.Source.Filename, submission.Source)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	cid := ctx.Params(":cid")
	pid := ctx.Params(":pid")

	// run go routine
	go runSubmission(filepath.Join("dataset", cid, pid), path, filepath.Join(path, submission.Source.Filename), filepath.Join(path, submissionId), fmt.Sprintf("%v", problem.TimeLimit), submission, &problem, &contest)

	ctx.Redirect("/contests/" + cid + "/" + pid)
}
