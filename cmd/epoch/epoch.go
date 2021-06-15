package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/frohwerk/deputy-backend/internal/epoch"
)

func main() {
	type stuff struct {
		Name  string       `json:"name,omitempty"`
		Value *epoch.Epoch `json:"value,omitempty"`
	}
	now := epoch.Epoch(time.Now())
	s := stuff{Name: "Test", Value: &now}
	buf, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(buf))
	}
	if true {
		return
	}
	t := time.Now()
	if len(os.Args) > 1 {
		fmt.Print(os.Args[1], " => ")
		var err error
		t, err = time.Parse("2006-01-02T15:04:05.000000Z07:00", os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	v := strconv.FormatInt(t.UnixNano()/1000, 10)
	l := len(v)
	fmt.Printf("%v.%v\n", v[:l-6], v[l-6:])
}
