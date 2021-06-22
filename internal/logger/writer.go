package logger

type logWriter struct {
	log logger
	lvl logLevel
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.log.log(w.lvl, string(p))
	return len(p), nil
}
