package rollout

func filter(slice []string, predicate func(string) bool) []string {
	res := make([]string, len(slice))
	copy(res, slice)
	for i := 0; i < len(res); {
		if !predicate(res[i]) {
			if last := len(res) - 1; last > 0 {
				res[i], res[last] = res[last], res[i]
				res = res[:last]
			} else {
				return []string{}
			}
		} else {
			i++
		}
	}
	return res
}
