package main

import "gopkg.in/macaron.v1"

func main() {
	m := macaron.Classic()
	m.Use(macaron.Renderer())

	m.Get("/", func(ctx *macaron.Context) {
		ctx.Data["Login"] = 0
		ctx.HTML(200, "index") // 200 is the response code.
	})

	m.Run()
}
