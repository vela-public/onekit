package treekit

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/layer"
)

func (mt *MsTree) diff(ctx *fasthttp.RequestCtx) (err error) {
	body := ctx.Request.Body()

	if len(body) == 0 {
		err = mt.HttpServiceView(ctx)
		return
	}

	var diff ServiceDiffInfo
	err = json.Unmarshal(body, &diff)
	if err != nil {
		return
	}

	//diff remove task by ids
	mt.RemoveByID(diff.Removes)

	errs := errkit.New()
	for _, entry := range diff.Updates {
		if e := mt.Register(entry.Name, entry.Chunk, func(c *MicoServiceConfig) {
			c.ID = entry.ID
			c.Hash = entry.Hash
			c.Dialect = entry.Dialect
		}); e != nil {
			errs.Try(entry.Name, e)
		}
	}

	if e := errs.Wrap(); e != nil {
		mt.Errorf(e.Error())
	}

	mt.Wakeup()
	return mt.HttpServiceView(ctx)
}

func (mt *MsTree) HttpServiceView(ctx *fasthttp.RequestCtx) error {
	_, e := ctx.Write(mt.Doc())
	return e
}

func (mt *MsTree) Define(route layer.RouterType) {
	_ = route.POST("/api/v1/agent/task/diff", route.Then(mt.diff))
	_ = route.POST("/api/v1/agent/task/status", route.Then(mt.HttpServiceView))
	_ = route.POST("/api/v1/arr/agent/task/status", route.Then(mt.HttpServiceView))

	_ = route.POST("/api/v1/agent/service/diff", route.Then(mt.diff))
	_ = route.POST("/api/v1/agent/service/status", route.Then(mt.HttpServiceView))
	_ = route.POST("/api/v1/arr/agent/service/status", route.Then(mt.HttpServiceView))
}
