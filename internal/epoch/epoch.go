package epoch

import (
	"fmt"
	"time"
)

const (
	second      = int64(time.Second)
	microsecond = int64(time.Microsecond)
)

type Epoch time.Time

func FromTime(t *time.Time) *Epoch {
	if t == nil {
		return nil
	}
	e := Epoch(*t)
	return &e
}

func (e *Epoch) MarshalText() ([]byte, error) {
	if e == nil {
		return nil, nil
	}
	t := time.Time(*e)
	sec := t.Unix()
	nsec := t.UnixNano() - sec*second
	return []byte(fmt.Sprintf("%v.%v", sec, nsec/microsecond)), nil
}

func (e *Epoch) MarshalJSON() ([]byte, error) {
	return e.MarshalText()
}
