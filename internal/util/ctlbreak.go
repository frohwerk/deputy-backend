package util

type ControlBreak struct {
	prev string
}

// IsBreak returns true when the *next* value is different from the previous invocation
func (cb *ControlBreak) IsBreak(next string) bool {
	defer func() { cb.prev = next }()
	return cb.prev != next
}
