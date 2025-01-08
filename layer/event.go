package layer

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	DISASTER string = "紧急"
	HIGH     string = "重要"
	MIDDLE   string = "次要"
	NOTICE   string = "普通"
)

type Event struct {
	MinionID   string    `json:"minio_id"`
	INet       string    `json:"inet"`
	Subject    string    `json:"subject"`
	RemoteAddr string    `json:"remote_addr"`
	RemotePort int       `json:"remote_port"`
	LocalAddr  string    `json:"local_addr"`
	LocalPort  int       `json:"local_port"`
	Region     string    `json:"region"`
	FromCode   string    `json:"from_code"`
	TypeOf     string    `json:"typeof"`
	User       string    `json:"user"`
	Auth       string    `json:"auth"`
	Message    string    `json:"msg"`
	Error      error     `json:"error"`
	Alert      bool      `json:"alert"`
	Level      string    `json:"level"`
	Time       time.Time `json:"time"`
}

func (e *Event) Byte() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}

func (e *Event) Log() *Event {
	return e
}

func (e *Event) Put() *Event {
	xEnv := LazyEnv()
	_ = xEnv.Push("/api/v1/broker/audit/event", e)
	return e
}

func NewEvent(typeof string) *Event {
	xEnv := LazyEnv()
	return &Event{
		MinionID: xEnv.ID(),
		INet:     xEnv.Inet(),
		Time:     time.Now(),
		Level:    NOTICE,
		TypeOf:   typeof,
	}
}

func Error(format string, v ...interface{}) *Event {
	ev := NewEvent("logger")
	ev.Subject = "发现错误"
	ev.Message = fmt.Sprintf(format, v...)
	return ev
}

func Debug(format string, v ...interface{}) *Event {
	ev := NewEvent("logger")
	ev.Subject = "调试信息"
	ev.Message = fmt.Sprintf(format, v...)
	return ev
}
