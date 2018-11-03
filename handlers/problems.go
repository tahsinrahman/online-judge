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

	"github.com/go-macaron/cache"
	redisCache "github.com/tahsinrahman/online-judge/cache"
	"github.com/tahsinrahman/online-judge/db"
	macaron "gopkg.in/macaron.v1"
)

//Dataset has a set of input and output files
type Dataset struct {
	Id          int64
	ProblemId   int64
	Label       string
	JudgeInput  string `json:"-" xorm:"longtext"`
	JudgeOutput string `json:"-" xorm:"longtext"`
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
	ProblemId int
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
	Description  string  `form:"description" xorm:"text"`
	Input        string  `form:"input" xorm:"longtext"`
	Output       string  `form:"output" xorm:"longtext"`
	SampleInput  string  `form:"sample_input" xorm:"longtext"`
	SampleOutput string  `form:"sample_output" xorm:"longtext"`
	TimeLimit    float64 `form:"timelimit"`
	MemoryLimit  int     `form:"memorylimit"`
	Notes        string  `form:"notes" xorm:"longtext"`
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

func addDataset(cid, pid string, problemId int64, dataset ProblemDataset, c cache.Cache) error {
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

		// remove old cache
		key := fmt.Sprintf("problem_%v_dataset", data.ProblemId)
		if err = c.Delete(key); err != nil {
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
func PostProblem(ctx *macaron.Context, c cache.Cache, problem Problem, dataset ProblemDataset) {
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

	// remove old cache
	cid := strconv.FormatInt(contest.Id, 10)
	pid := strconv.Itoa(problem.ProblemId)

	err = addDataset(cid, pid, problem.Id, dataset, c)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// update contest in db
	if _, err = db.Engine.Id(contest.Id).Update(&contest); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// remove old cache
	key := fmt.Sprintf("contest_%v_problems", cid)
	if err = c.Delete(key); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	key = "contest_list"
	if err = c.Delete(key); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	key = fmt.Sprintf("contest_%v", contest.Id)
	if err = c.Delete(key); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	//finally redirec to to dashboard
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
func UpdateProblemDescripton(ctx *macaron.Context, c cache.Cache, problem Problem) {
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

	oldProblem := ctx.Data["Problem"].(Problem)

	// update problem in db
	if _, err = db.Engine.Id(oldProblem.Id).Update(&problem); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	// update problem in cache
	key := fmt.Sprintf("contest_%v_problem_%v", cid, pid)
	if err = c.Delete(key); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
	key = fmt.Sprintf("contest_%v_problems", cid)
	if err = c.Delete(key); err != nil {
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

func DeleteTest(ctx *macaron.Context, c cache.Cache) {
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

	// remove old cache
	problem := ctx.Data["Problem"].(Problem)
	key := fmt.Sprintf("problem_%v_dataset", problem.Id)
	if err = c.Delete(key); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
}

func AddNewTest(ctx *macaron.Context, c cache.Cache, dataset ProblemDataset) {
	if len(dataset.JudgeInput) == 0 {
		ctx.Resp.Write([]byte("nothing found"))
		return
	}

	problem := ctx.Data["Problem"].(Problem)

	err := addDataset(ctx.Params(":cid"), ctx.Params(":pid"), problem.Id, dataset, c)
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}
}

func updateSubmissionStatus(c cache.Cache, submission *Submission, contest *Contest, problem *Problem, oldSubmission string) error {
	rank := Rank{
		ContestId: contest.Id,
		ProblemId: problem.ProblemId,
		UserName:  submission.UserName,
	}
	has, err := db.Engine.Get(&rank)
	if err != nil {
		return err
	}

	if has {
		if rank.Score < submission.Points {
			rank.Score = submission.Points
			rank.Penalty = int64(submission.Time.Sub(contest.ContestStartTime).Minutes()) + int64(rank.Tries*20)
		}
		rank.Tries++

		if int(rank.Score) == problem.MaxPoint {
			rank.Status = 1
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
			rank.Penalty = int64(submission.Time.Sub(contest.ContestStartTime).Minutes())
		}
		if int(rank.Score) == problem.MaxPoint {
			rank.Status = 1
		}

		// insert
		_, err := db.Engine.InsertOne(&rank)
		if err != nil {
			return err
		}
	}

	if _, err = db.Engine.Id(submission.Id).Update(submission); err != nil {
		return err
	}

	// update cache
	// first remove old cache
	key := fmt.Sprintf("submission_%v", submission.Id)
	if err = c.Delete(key); err != nil {
		panic(err)
	}

	key = fmt.Sprintf("contest_%v_submissions", contest.Id)
	if err := db.Client.SRem(key, oldSubmission).Err(); err != nil {
		panic(err)
	}
	key = fmt.Sprintf("submissions_user_%v_problem_%v", submission.UserName, problem.Id)
	if err = db.Client.SRem(key, oldSubmission).Err(); err != nil {
		panic(err)
	}

	contestSubmission := ContestSubmission{Submission: *submission, Name: problem.Name, ContestId: contest.Id}

	b, err := json.Marshal(contestSubmission)
	if err != nil {
		panic(err)
	}

	// now update new cache
	key = fmt.Sprintf("contest_%v_submissions", contest.Id)
	if db.Client.Exists(key).Val() > int64(0) {
		if err = db.Client.SAdd(key, string(b)).Err(); err != nil {
			return err
		}
	}
	key = fmt.Sprintf("submissions_user_%v_problem_%v", submission.UserName, problem.Id)
	if db.Client.Exists(key).Val() > int64(0) {
		if err = db.Client.SAdd(key, string(b)).Err(); err != nil {
			return err
		}
	}

	key = fmt.Sprintf("submission_%v", submission.Id)
	return c.Delete(key)

	//return redisCache.AddToList(key, &submission)
}

func compile(c cache.Cache, compileCmd string, submission *Submission, contest *Contest, problem *Problem) bool {
	args := strings.Split(compileCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)
	var buf bytes.Buffer
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		// remove old cache
		key := fmt.Sprintf("submission_%v", submission.Id)
		if err = c.Delete(key); err != nil {
			//FIXME
			panic(err)
		}
		b, err := json.Marshal(submission)
		if err != nil {
			panic(err)
		}

		/*
			key = fmt.Sprintf("contest_%v_submissions", contest.Id)
			if err = db.Client.SRem(key, string(b)).Err(); err != nil {
				panic(err)
			}
			key = fmt.Sprintf("submissions_user_%v_problem_%v", submission.UserName, problem.Id)
			if err = db.Client.SRem(key, string(b)).Err(); err != nil {
				panic(err)
			}
			key = fmt.Sprintf("submission_%v", submission.Id)
			if err = c.Delete(key); err != nil {
				panic(err)
			}
		*/

		submission.Status = "compilation error"
		err = updateSubmissionStatus(c, submission, contest, problem, string(b))
		if err != nil {
			//FIXME
			panic(err)
		}
		return false
	}
	return true
}

type DatasetVerdict struct {
	Id           int64
	SubmissionId int64
	DatasetId    int64
	Verdict      string
	CPU          string
}

func run(c cache.Cache, cid, pid string, submission *Submission, contest *Contest, problem *Problem) {
	// remove old cache
	oldSubmission, err := json.Marshal(submission)
	if err != nil {
		panic(err)
	}

	// now run for all tests
	key := fmt.Sprintf("problem_%v_dataset", submission.ProblemId)
	dbSession := db.Engine.Where("problem_id = ?", submission.ProblemId)
	var data []Dataset

	if _, err := redisCache.FindObject(c, key, &data, dbSession, false); err != nil {
		panic(err)
	}

	dataPath := filepath.Join("dataset", cid, pid)

	total, maxPossible := 0, 0
	for _, dataset := range data {
		submissionId := strconv.FormatInt(submission.Id, 10)
		datasetId := strconv.FormatInt(dataset.Id, 10)

		cmdRun := exec.Command("/bin/bash", "scripts/runTest.sh", dataPath, submissionId, datasetId, fmt.Sprintf("%v", problem.TimeLimit), submission.Language, submission.Source.Filename)

		var buf, stderr bytes.Buffer
		cmdRun.Stdout = &buf
		cmdRun.Stderr = &stderr

		if err := cmdRun.Run(); err == nil {
			total += dataset.Weight
		}

		verdict := buf.String()

		if verdict == "139\n" {
			verdict = "Runtime Error"
			if len(submission.Status) > 0 {
				submission.Status = verdict
			}
		} else if verdict == "WA\n" {
			verdict = "Wrong Answer"
			if len(submission.Status) > 0 {
				submission.Status = verdict
			}
		} else if verdict != "" {
			verdict = "Time Limit Exceeded"
			if len(submission.Status) > 0 {
				submission.Status = verdict
			}
		} else {
			verdict = "Accepted"
		}

		maxPossible += dataset.Weight

		// update verdict for this dataset
		cputime := strings.Split(strings.Split(stderr.String(), "\n")[1], "\t")[1]
		_, err := db.Engine.Insert(&DatasetVerdict{SubmissionId: submission.Id, DatasetId: dataset.Id, Verdict: verdict, CPU: cputime})
		if err != nil {
			panic(err)
		}
		key := fmt.Sprintf("submission_%v_verdicts", submission.Id)
		if err := c.Delete(key); err != nil {
			panic(err)
		}
	}

	if total == maxPossible {
		submission.Status = "Accepted"
	}

	submission.Points = float64(problem.MaxPoint) * float64(total) / float64(maxPossible)

	if err := updateSubmissionStatus(c, submission, contest, problem, string(oldSubmission)); err != nil {
		panic(err)
	}
}

func runSubmission(c cache.Cache, cid, pid, sourcePath, filename, execName string, submission *Submission, problem *Problem, contest *Contest) {
	switch submission.Language {
	case "java":
		compileCmd := fmt.Sprintf("javac %s", filename)
		if !compile(c, compileCmd, submission, contest, problem) {
			return
		}
		run(c, cid, pid, submission, contest, problem)
	case "c++":
		compileCmd := fmt.Sprintf("g++ -O3 -std=c++14 -o %s %s", execName, filename)
		if !compile(c, compileCmd, submission, contest, problem) {
			return
		}
		run(c, cid, pid, submission, contest, problem)
	case "c":
		compileCmd := fmt.Sprintf("gcc -O3 -o %s %s", execName, filename)
		if !compile(c, compileCmd, submission, contest, problem) {
			return
		}
		run(c, cid, pid, submission, contest, problem)
	case "python2":
		compileCmd := fmt.Sprintf("python2 -m py_compile %s", filename)
		if !compile(c, compileCmd, submission, contest, problem) {
			return
		}
		run(c, cid, pid, submission, contest, problem)
	case "python3":
		compileCmd := fmt.Sprintf("python3 -m py_compile %s", filename)
		if !compile(c, compileCmd, submission, contest, problem) {
			return
		}
		run(c, cid, pid, submission, contest, problem)
	default:
		return
	}
}

//route: /contests/:cid/:pid POST
//submit problem if eligible and logged in
func SubmitProblem(submission Submission, ctx *macaron.Context, c cache.Cache) {
	if ctx.Data["Username"] == nil {
		ctx.Resp.Write([]byte("login"))
		return
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

	// add to cache
	key := fmt.Sprintf("contest_%v_submissions", contest.Id)
	if db.Client.Exists(key).Val() != int64(0) {
		if err := redisCache.AddToList(key, &submission); err != nil {
			ctx.Resp.Write([]byte(err.Error()))
			return
		}
	}
	key = fmt.Sprintf("submissions_user_%v_problem_%v", submission.UserName, problem.Id)
	if db.Client.Exists(key).Val() != int64(0) {
		if err := redisCache.AddToList(key, &submission); err != nil {
			ctx.Resp.Write([]byte(err.Error()))
			return
		}
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
	go runSubmission(c, cid, pid, path, filepath.Join(path, submission.Source.Filename), filepath.Join(path, submissionId), &submission, &problem, &contest)

	ctx.Redirect("/contests/" + cid + "/" + pid)
}

func GetSubmission(ctx *macaron.Context, c cache.Cache) {
	var verdicts []DatasetVerdict
	key := fmt.Sprintf("submission_%v_verdicts", ctx.Params("id"))
	dbSession := db.Engine.Where("submission_id = ?", ctx.Params("id"))

	if _, err := redisCache.FindObject(c, key, &verdicts, dbSession, false); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	subId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	key = fmt.Sprintf("submission_%v", subId)
	submission := Submission{Id: int64(subId)}

	if _, err := redisCache.FindObject(c, key, &submission, nil, true); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	key = fmt.Sprintf("problem_%v", submission.ProblemId)
	problem := Problem{Id: submission.ProblemId}

	if _, err = redisCache.FindObject(c, key, &problem, nil, true); err != nil {
		ctx.Resp.Write([]byte(err.Error()))
		return
	}

	ctx.Data["Submission"] = submission
	ctx.Data["Problem"] = problem
	ctx.Data["Verdicts"] = verdicts

	ctx.HTML(200, "submission")
}
