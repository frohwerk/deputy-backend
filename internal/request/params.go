package request

import "strconv"

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
