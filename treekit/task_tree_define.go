package treekit

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/layer"
)

func (t *TaskTree) PushTask(ctx *fasthttp.RequestCtx) error {
	body := ctx.Request.Body()
	config := &TaskConfig{}

	err := json.Unmarshal(body, config)
	if err != nil {
		return err
	}

	tas := NewTask(t, config)

	if e := tas.do(); e != nil {
		return e
	}

	return nil
}

func (t *TaskTree) Define(route layer.RouterType) {
	_ = route.POST("/api/v1/agent/task/push", route.Then(t.PushTask))
}
