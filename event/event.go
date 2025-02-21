package event

import (
	"encoding/json"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/todo"
	"time"
)

const (
	DISASTER string = "紧急"
	HIGH     string = "重要"
	MIDDLE   string = "次要"
	NOTICE   string = "普通"
)

type Metadata map[string]any

func (m Metadata) Text() string {
	text, _ := json.Marshal(m)
	return cast.B2S(text)
}

type Event struct {
	Time     time.Time `json:"time"`
	MinionID string    `json:"minio_id"`
	INet     string    `json:"inet"`
	Subject  string    `json:"subject"`
	FromCode string    `json:"from_code"`
	TypeOf   string    `json:"typeof"`
	Message  string    `json:"msg"`
	Alert    bool      `json:"alert"`
	Level    string    `json:"level"`
	Metadata Metadata  `json:"metadata"` /* user , auth , remote_addr , remote_port , local_addr , local_port , region etc */

	//内部变量
	private struct {
		Env layer.Environment
	}
}

func (e *Event) Byte() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}

func (e *Event) Set(key string, val any) {
	if e.Metadata == nil {
		e.Metadata = Metadata{
			key: val,
		}
		return
	}
	e.Metadata[key] = val
}

func (e *Event) Text() string {
	text := fmt.Sprintf("%s %s %s %s %s %s %s %s",
		e.MinionID,
		e.INet,
		todo.IF(e.Subject == "", "-", e.Subject),
		todo.IF(e.FromCode == "", "-", e.FromCode),
		todo.IF(e.TypeOf == "", "-", e.TypeOf),
		todo.IF(e.Message == "", "-", e.Message),
		todo.IF(e.Level == "", "-", e.Level),
		todo.IF(len(e.Metadata) == 0, "-", e.Metadata.Text()),
	)
	return text
}

func (e *Event) Error() *Event {
	if e.private.Env == nil {
		return e
	}
	e.private.Env.Logger().Error(e.Text())
	return e
}

func (e *Event) Debug() *Event {
	if e.private.Env == nil {
		return e
	}
	e.private.Env.Logger().Debug(e.Text())
	return e
}

func (e *Event) Info(logger layer.LoggerType) *Event {
	logger.Info(e.Text())
	return e
}

func (e *Event) Report() *Event {
	if e.private.Env == nil {
		return e
	}

	err := e.private.Env.Transport().Push("/api/v1/broker/audit/event", e)
	if err != nil {
		e.private.Env.Logger().Error(err)
	}

	return e
}

func NewEvent(xEnv layer.Environment, typeof string) *Event {
	ev := &Event{
		MinionID: xEnv.ID(),
		INet:     xEnv.IP(),
		Time:     time.Now(),
		Level:    NOTICE,
		TypeOf:   typeof,
		Metadata: Metadata{},
	}
	ev.private.Env = xEnv
	return ev
}

func Error(xEnv layer.Environment, format string, v ...interface{}) *Event {
	ev := NewEvent(xEnv, "logger")
	ev.Subject = "发现错误"
	ev.Message = fmt.Sprintf(format, v...)
	return ev
}

func Debug(xEnv layer.Environment, format string, v ...interface{}) *Event {
	ev := NewEvent(xEnv, "logger")
	ev.Subject = "调试信息"
	ev.Message = fmt.Sprintf(format, v...)
	return ev
}

func Trace(xEnv layer.Environment, format string, v ...interface{}) *Event {
	ev := NewEvent(xEnv, "logger")
	ev.Subject = "提示信息"
	ev.Message = fmt.Sprintf(format, v...)
	return ev
}
