package util

func MaxInt(a int, values ...int) int {
	for _, i := range values {
		if i > a {
			a = i
		}
	}
	return a
}
