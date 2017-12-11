package main

import (
	"gopkg.in/macaron.v1"
)

func main() {
	m := macaron.Classic()
	m.Use(macaron.Renderer())

	m.Get("/", func(ctx *macaron.Context) {
		ctx.Data["Login"] = 0
		ctx.HTML(200, "index")
	})

	m.Get("/signup", func(ctx *macaron.Context) {
		ctx.HTML(200, "signup")
	})
	m.Post("/signup", func(ctx *macaron.Context) {
		//insert to DB
		ctx.Redirect(ctx.Req.URL.Host, 301)
	})

	m.Get("/signin", func(ctx *macaron.Context) {
		ctx.HTML(200, "signin")
	})
	m.Post("/signin", func(ctx *macaron.Context) {
		//insert to DB
		ctx.Redirect(ctx.Req.URL.Host, 301)
	})

	m.Get("/signout", func(ctx *macaron.Context) {
		ctx.HTML(200, "signout")
	})

	m.Run(8000)
}
