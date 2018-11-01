package handlers

import macaron "gopkg.in/macaron.v1"

func PD(ctx *macaron.Context) {
	ctx.HTML(200, "plagiarism")
}
