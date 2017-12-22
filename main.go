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

	//middlewares
	m.Use(macaron.Renderer())              //for html rendering
	m.Use(middlewares.CheckAuthentication) //check cookie

	//handlers for user registration, login and logout
	m.Get("/", handlers.GetHome)
	m.Get("/signup", handlers.GetSignUp)
	m.Post("/signup", binding.Bind(handlers.Users{}), handlers.PostSignUp)
	m.Get("/signin", handlers.GetSignIn)
	m.Post("/signin", binding.Bind(handlers.Users{}), handlers.PostSignIn)
	m.Get("/signout", handlers.GetSignOut)

	//starting the server
	m.Run()
}
