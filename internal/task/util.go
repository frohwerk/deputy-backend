package task

func StartAll(tasks map[string]Task) {
	for _, task := range tasks {
		go task.Start()
	}
}

func StopAll(tasks map[string]Task) {
	for _, task := range tasks {
		task.Stop()
	}
}

func WaitAll(tasks map[string]Task) {
	for _, task := range tasks {
		task.Wait()
	}
}
