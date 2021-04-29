package predicates

import "strings"

func Prefix(prefixes ...string) Predicate {
	return func(path string) bool {
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		return false
	}
}
