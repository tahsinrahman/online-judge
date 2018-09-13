package main

import (
	"github.com/go-macaron/binding"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tahsinrahman/online-judge/db"
	"github.com/tahsinrahman/online-judge/handlers"
	"github.com/tahsinrahman/online-judge/middlewares"
	"gopkg.in/macaron.v1"
)

func main() {
	//starts database engine
	db.StartEngine()

	//macaron engine
	m := macaron.Classic()

	//middlewares for every route
	m.Use(macaron.Renderer())              //for html rendering
	m.Use(middlewares.CheckAuthentication) //check cookie

	//handlers for user registration, login and logout TODO: secret key based cookie
	m.Get("/", handlers.GetHome)                                           //done
	m.Get("/signup", handlers.GetSignUp)                                   //done
	m.Post("/signup", binding.Bind(handlers.Users{}), handlers.PostSignUp) //done
	m.Get("/signin", handlers.GetSignIn)                                   //done
	m.Post("/signin", binding.Bind(handlers.Users{}), handlers.PostSignIn) //done
	m.Get("/signout", handlers.GetSignOut)                                 //done

	//	m.Get("/upload", handlers.GetUpload) //done
	//	m.Post("/upload", handlers.Upload)   //done

	//handlers for contests
	m.Group("/contests", func() {
		m.Get("/", handlers.GetAllContests) //done

		m.Get("/new", handlers.GetNewContest)                                  //done
		m.Post("/new", binding.Bind(handlers.Contest{}), handlers.PostContest) //done

		m.Group("/:cid", func() {
			m.Get("/", handlers.GetDashboard) //done

			m.Group("", func() {
				m.Delete("/", handlers.DeleteContest) //TODO:

				m.Get("/update", handlers.GetUpdateContest)                              //done
				m.Post("/update", binding.Bind(handlers.Contest{}), handlers.PutContest) //done

				m.Get("/new", handlers.GetNewProblem)                                                                             //done
				m.Post("/new", binding.Bind(handlers.Problem{}), binding.MultipartForm(handlers.Dataset{}), handlers.PostProblem) //TODO:db
			}, middlewares.CheckManager)

			m.Group("/:pid", func() {
				m.Get("/", handlers.GetProblem)
				m.Get("/update", handlers.UpdateProblem)
				m.Post("/update", binding.Bind(handlers.Problem{}), binding.MultipartForm(handlers.Dataset{}), handlers.PutPostProblem)
				m.Delete("/:pid", handlers.DeleteProblem)
				m.Post("/:pid/submit", handlers.SubmitProblem)
			}, middlewares.CheckProblem) //need to add middleware to check if problem exists

			m.Get("/rank", handlers.GetRank)
			m.Get("/allsubmissions", handlers.GetAllSubmissions)
			m.Get("/mysubmissions", handlers.GetMySubmissions)
		}, middlewares.CheckContestExistance)
	})

	handlers.Init()

	//starting the server
	m.Run()
}
