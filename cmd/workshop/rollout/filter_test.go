package rollout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	contains := func(slice []string, v string) bool {
		for _, s := range slice {
			if s == v {
				return true
			}
		}
		return false
	}

	t.Run("original unmodified", func(t *testing.T) {
		original := []string{"a", "b", "c", "e", "f"}
		allowed := []string{"b", "d", "c"}
		result := filter(original, func(s string) bool { return contains(allowed, s) })
		assert.Equal(t, []string{"a", "b", "c", "e", "f"}, original)
		assert.Equal(t, []string{"b", "d", "c"}, allowed)
		assert.ElementsMatch(t, []string{"b", "c"}, result)
	})

}
