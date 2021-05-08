package request

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
