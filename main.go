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
	m.SetDefaultCookieSecret("tahsin")
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

		m.Get("/new", middlewares.CheckAdmin, handlers.GetNewContest)                                  //done
		m.Post("/new", middlewares.CheckAdmin, binding.Bind(handlers.Contest{}), handlers.PostContest) //done

		m.Group("/:cid", func() {
			m.Get("/", handlers.GetDashboard)    //done
			m.Get("/auth", handlers.ContestAuth) //done
			m.Get("/rank", handlers.GetRank)
			m.Get("/submissions/:id", handlers.GetSubmission)

			m.Group("", func() {
				m.Get("/update", handlers.GetUpdateContest)                              //done
				m.Post("/update", binding.Bind(handlers.Contest{}), handlers.PutContest) //done

				m.Get("/new", handlers.GetNewProblem)                                                                                    //done
				m.Post("/new", binding.Bind(handlers.Problem{}), binding.MultipartForm(handlers.ProblemDataset{}), handlers.PostProblem) //TODO:db
			}, middlewares.CheckManager)

			m.Group("/:pid", func() {
				m.Get("/", middlewares.CheckStartTime, middlewares.AddSubmissions, handlers.GetProblem)
				//m.Delete("/", handlers.DeleteProblem)
				//m.Post("/update", binding.Bind(handlers.Problem{}), binding.MultipartForm(handlers.ProblemDataset{}), handlers.PutPostProblem)

				m.Get("/dload/:type/:id", middlewares.CheckManager, handlers.DownloadTest)
				m.Get("/tests", middlewares.CheckManager, middlewares.AddTests, handlers.GetList)
				m.Post("/tests", middlewares.CheckManager, binding.Bind(handlers.ProblemDataset{}), handlers.AddNewTest)
				m.Post("/submit", middlewares.CheckStartTime, middlewares.CheckEndTime, binding.MultipartForm(handlers.Submission{}), handlers.SubmitProblem)

				m.Group("/update", func() {
					m.Get("/", handlers.UpdateProblem)
					m.Post("/description", binding.Bind(handlers.Problem{}), handlers.UpdateProblemDescripton)
					m.Delete("/:id", handlers.DeleteTest)
				}, middlewares.CheckManager, middlewares.AddTests)
			}, middlewares.CheckProblem) //need to add middleware to check if problem exists

		}, middlewares.CheckContestExistance, middlewares.AddContestPermission)
	})

	handlers.Init()

	//starting the server
	m.Run()
}
