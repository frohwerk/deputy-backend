package epoch

import (
	"fmt"
	"strconv"
	"strings"
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

func ParseTime(s string) (time.Time, error) {
	parts := strings.SplitN(s, ".", 2)
	secs, err := strconv.ParseInt(parts[0], 10, 64)
	time.Parse(time.UnixDate, "")
	if err != nil {
		return time.Time{}, err
	}
	nsecs, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(secs, nsecs*microsecond).In(time.UTC), nil
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
