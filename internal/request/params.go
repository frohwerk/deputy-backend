package request

import (
	"fmt"
	"strconv"
	"time"

	"github.com/frohwerk/deputy-backend/internal/epoch"
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
		time, err := epoch.ParseTime(v[0])
		if err != nil {
			fmt.Printf("invalid time parameter '%s': %s\n", v[0], err)
			return nil, false
		}
		return &time, true
	}
	return nil, false
}
