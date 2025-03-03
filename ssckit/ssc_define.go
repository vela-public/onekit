package ssckit

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

func (app *Application) startup() {
	r := app.Transport().R()
	_ = r.POST("/api/v1/agent/startup", r.Then(func(ctx *fasthttp.RequestCtx) error {
		cfg := Config{}
		dat := ctx.Request.Body()
		err := json.Unmarshal(dat, &cfg)
		if err != nil {
			return err
		}
		app.private.Logger.Apply(cfg.Logger)
		app.config = &cfg
		return nil
	}))
}
