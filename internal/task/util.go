package task

func StartAll(tasks []Task) {
	for _, task := range tasks {
		go task.Start()
	}
}

func StopAll(tasks []Task) {
	for _, task := range tasks {
		task.Stop()
	}
}

func WaitAll(tasks []Task) {
	for _, task := range tasks {
		task.Wait()
	}
}
