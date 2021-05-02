package task

type state int

const (
	Created = state(iota)
	Running = state(iota)
	Stopped = state(iota)
)
