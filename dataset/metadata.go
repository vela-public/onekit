package dataset

import (
	"fmt"
	"strconv"
	"strings"
)

/*
text(100)
*/

type Attribute struct {
	Offset int
	Size   int
	Align  int
	Name   string
	Kind   string
	Text   struct {
		Len  int
		Data []byte
	}
}

type Structural struct {
	offset     int
	max        int
	total      int
	attributes []Attribute
}

func (s *Structural) AlignOf(att *Attribute) {
	if att.Align == 0 {
		return
	}
	s.offset = (s.offset + att.Align - 1) & ^(att.Align - 1)
	att.Offset = s.offset
	s.offset += att.Size

	/*
		offset = alignUp(offset, meta.Align)
		meta.Offset = offset
		metas = append(metas, meta)
		offset += meta.Size
		if meta.Align > maxAlign {
			maxAlign = meta.Align
		}
	*/
}

func (s *Structural) Layout(v string) error {
	dat := strings.Replace(v, " ", "", -1)
	mark := strings.SplitN(dat, ":", 2)
	if len(mark) != 2 {
		return fmt.Errorf("error desc must name:kind")
	}

	name := mark[0]
	kind := mark[1]
	if len(name) == 0 {
		return fmt.Errorf("name is empty")
	}
	if len(kind) == 0 {
		return fmt.Errorf("kind is empty")
	}

	att := Attribute{
		Name: name,
	}

	switch {
	case kind == "int32":
		att.Size = 4
		att.Align = 4
		att.Kind = "int32"
	case kind == "bool":
		att.Size = 1
		att.Align = 1
		att.Kind = "bool"
	case kind == "int":
		att.Size = 4
		att.Align = 4
		att.Kind = "int"
	case kind == "int64":
		att.Size = 8
		att.Align = 8
		att.Kind = "int64"
	case strings.HasSuffix(kind, "text("):
		att.Kind = "text"
		text := strings.TrimPrefix(kind, "text(")
		text = strings.TrimSuffix(text, ")")
		if len(text) == 0 {
			return fmt.Errorf("text length is empty")
		}

		n, err := strconv.Atoi(text)
		if err != nil {
			return err
		}
		att.Size = n
		att.Align = 1
		att.Text.Len = n
	default:
		return fmt.Errorf("unsupported kind %s", kind)
	}

	if att.Align == 0 {
		return fmt.Errorf("unsupport kind %s case align = 0", kind)
	}

	s.AlignOf(&att)
	s.attributes = append(s.attributes, att)
	if s.max < att.Align {
		s.max = att.Align
	}

	return nil
}

func NewMetadata() *Structural {
	return &Structural{
		offset: 0,
		max:    1,
	}
}
