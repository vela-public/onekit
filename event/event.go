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

type Metadata map[string]string

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
}

func (e *Event) Byte() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
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

func (e *Event) Error(logger layer.LoggerType) *Event {
	logger.Error(e.Text())
	return e
}

func (e *Event) Debug(logger layer.LoggerType) *Event {
	logger.Debug(e.Text())
	return e
}

func (e *Event) Info(logger layer.LoggerType) *Event {
	logger.Info(e.Text())
	return e
}

func (e *Event) Put(transport layer.Transport) *Event {
	_ = transport.Push("/api/v1/broker/audit/event", e)
	return e
}

func NewEvent(xEnv layer.Environment, typeof string) *Event {
	return &Event{
		MinionID: xEnv.ID(),
		INet:     xEnv.IP(),
		Time:     time.Now(),
		Level:    NOTICE,
		TypeOf:   typeof,
	}
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
