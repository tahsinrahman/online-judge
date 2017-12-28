package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	macaron "gopkg.in/macaron.v1"
)

type Dataset struct {
	JudgeInput  []*multipart.FileHeader `form:"input_data[]"`
	JudgeOutput []*multipart.FileHeader `form:"output_data[]"`
}

type Problem struct {
	ProblemId    int
	ContestId    int
	OrderId      string
	Name         string  `form:"name"`
	Description  string  `form:"description"`
	Input        string  `form:"input"`
	Output       string  `form:"output"`
	SampleInput  string  `form:"sample_input"`
	SampleOutput string  `form:"sample_output"`
	TimeLimit    float64 `form:"timelimit"`
	Notes        string  `form:"notes`
}

//route: /contests/:cid/new GET
func GetNewProblem(ctx *macaron.Context) {
	ctx.HTML(200, "new_problem")
}

//route: /contests/:cid/new POST
func PostNewProblem(ctx *macaron.Context, problem Problem, dataset Dataset) {
	//	fmt.Println(problem)
	//	fmt.Println(dataset)
	//TODO: upload data to db
	fmt.Println((len(dataset.JudgeInput)))
	fmt.Println((len(dataset.JudgeOutput)))

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	/*
		file, err := judgeData.JudgeInput.Open()
		if err != nil {
			fmt.Println(err)
			return
		}
		data := make([]byte, 100)
		_, err = file.Read(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(data))
	*/
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
