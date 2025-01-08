package taskit

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/layer"
)

func (t *Tree) diff(ctx *fasthttp.RequestCtx) (err error) {
	body := ctx.Request.Body()

	if len(body) == 0 {
		err = t.HttpTaskView(ctx)
		return
	}

	var diff TaskDiffInfo
	err = json.Unmarshal(body, &diff)
	if err != nil {
		return
	}

	//diff remove task by ids
	t.RemoveByID(diff.Removes)

	errs := errkit.New()
	for _, entry := range diff.Updates {
		if e := t.Register(entry.Name, entry.Chunk, func(c *Config) {
			c.ID = entry.ID
			c.Hash = entry.Hash
			c.Dialect = entry.Dialect
		}); e != nil {
			errs.Try(entry.Name, e)
		}
	}

	if e := errs.Wrap(); e != nil {
		fmt.Println(e.Error())
	}

	t.Wakeup()
	return t.HttpTaskView(ctx)
}

func (t *Tree) HttpTaskView(ctx *fasthttp.RequestCtx) error {
	_, e := ctx.Write(t.Doc())
	return e
}

func (t *Tree) Define(route layer.RouterType) {
	_ = route.POST("/api/v1/agent/task/diff", route.Then(t.diff))
	_ = route.POST("/api/v1/agent/task/status", route.Then(t.HttpTaskView))
	_ = route.POST("/api/v1/arr/agent/task/status", route.Then(t.HttpTaskView))
}
