package image

import (
	"testing"
)

func TestSplitReverseN(t *testing.T) {
	type testcase struct {
		s        string
		sep      string
		n        int
		expected []string
	}
	t.Run("Basic test", func(t *testing.T) {
		tests := []testcase{
			{"", "", 2, []string{""}},
			{"", ":", 2, []string{""}},
			{"abc@def", "", 2, []string{"abc@def"}},
			{"abc@def", "@", 2, []string{"abc", "def"}},
			{"abc@def@ghi", "@", 2, []string{"abc@def", "ghi"}},
			{"abc@def:ghi", ":", 2, []string{"abc@def", "ghi"}},
		}
		for _, tc := range tests {
			t.Logf(`testing: splitReverseN("%v", "%v", %v)`, tc.s, tc.sep, tc.n)
			actual := splitReverseN(tc.s, tc.sep, tc.n)
			if !equal(tc.expected, actual) {
				t.Errorf(`expected splitReverseN("%v", "%v", %v) to return %v, but it returned %v`, tc.s, tc.sep, tc.n, tc.expected, actual)
			}
		}
	})
}

func equal(a []string, b []string) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	case len(a) != len(b):
		return false
	default:
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}
