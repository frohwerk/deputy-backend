package job

type Runner interface {
	Run(Params) error
}

type Job struct {
	Out Output
}
