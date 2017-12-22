package handlers

		//homepage
		func GetHome(ctx *macaron.Context) {
			ctx.HTML(200, "index")
		})

