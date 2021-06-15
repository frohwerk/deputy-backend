package request

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func BooleanParam(queryParams map[string][]string, name string) bool {
	v, ok := queryParams[name]
	return ok && (len(v) == 0 || v[0] == "")
}

func StringParam(queryParams map[string][]string, name string) (string, bool) {
	if v, exists := queryParams[name]; exists && len(v) > 0 {
		return v[0], true
	}
	return "", false
}

func FloatParam(queryParams map[string][]string, name string) (float64, bool) {
	if v, exists := queryParams[name]; exists && len(v) > 0 {
		value, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			return 0, false
		}
		return value, true
	}
	return 0, false
}

func TimeParam(queryParams map[string][]string, name string) (*time.Time, bool) {
	if v, exists := queryParams[name]; exists && len(v) > 0 {
		parts := strings.SplitN(v[0], ".", 2)
		sec, err := strconv.ParseInt(parts[0], 10, 64)
		msec := int64(0)
		if err == nil && len(parts) == 2 {
			msec, err = strconv.ParseInt(parts[1], 10, 64)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing TimeParam: %v\n", v[0])
			return nil, false
		}
		time := time.Unix(sec, msec*1000).In(time.UTC)
		return &time, true
	}
	return nil, false
}
