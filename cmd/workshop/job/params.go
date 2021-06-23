package job

type Params map[string][]string

func (p Params) Get(s string) string {
	if v, exists := p[s]; exists {
		if len(v) == 0 {
			return ""
		}
		return v[0]
	}
	return ""
}
