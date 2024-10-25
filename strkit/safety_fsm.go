package strkit

import (
	"bytes"
	"github.com/vela-public/onekit/cast"
	"strconv"
	"strings"
)

func (fsm *SafetyFSM) Last() *MaskTag {
	if sz := len(fsm.Mask); sz > 0 {
		return fsm.Mask[sz-1]
	}
	return nil
}

func (fsm *SafetyFSM) Norm() string {
	sz := len(fsm.Mask)
	if len(fsm.MaskText.norm) != 0 {
		return cast.B2S(fsm.MaskText.norm)
	}
	var buf bytes.Buffer

	for i := 0; i < sz; i++ {
		m := fsm.Mask[i]
		buf.WriteString(m.Tag)
		n := m.To - m.From
		if n > 1 {
			buf.WriteString(strconv.Itoa(n))
		}
	}
	fsm.MaskText.norm = buf.Bytes()
	return cast.B2S(fsm.MaskText.norm)
}

func (fsm *SafetyFSM) Detail() string {
	sz := len(fsm.Mask)
	if sz == 0 {
		return ""
	}

	if len(fsm.MaskText.detail) != 0 {
		return cast.B2S(fsm.MaskText.detail)
	}
	var buf bytes.Buffer

	for i := 0; i < sz; i++ {
		m := fsm.Mask[i]
		buf.WriteString(m.Tag)
		if m.To-m.From > 1 {
			buf.WriteByte('[')
			buf.WriteString(strconv.Itoa(m.From))
			buf.WriteByte(':')
			buf.WriteString(strconv.Itoa(m.To))
			buf.WriteByte(']')
		}
	}
	fsm.MaskText.detail = buf.Bytes()
	return cast.B2S(fsm.MaskText.detail)
}

func (fsm *SafetyFSM) Simplicity() string {
	sz := len(fsm.Mask)
	if sz == 0 {
		return ""
	}

	if len(fsm.MaskText.simple) != 0 {
		return cast.B2S(fsm.MaskText.simple)
	}
	var buf bytes.Buffer

	for i := 0; i < sz; i++ {
		m := fsm.Mask[i]
		buf.WriteString(m.Tag)
	}

	fsm.MaskText.simple = buf.Bytes()
	return cast.B2S(fsm.MaskText.simple)

}

func (fsm *SafetyFSM) BadText() string {
	return strings.Join(fsm.Bad, ",")
}
