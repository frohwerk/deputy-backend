package main

import (
	"sync"

	"github.com/frohwerk/deputy-backend/internal/task"
)

type controller struct {
	sync.Mutex
	closed bool
	tasks  map[string]task.Task
}

func (c *controller) Start(id string, t task.Task) {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return
	}
	c.tasks[id] = t
	go t.Start()
}

func (c *controller) Restart(id string) {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return
	}
	task := c.tasks[id]
	task.Stop()
	go task.Start()
}

func (c *controller) Remove(id string) {
	c.Lock()
	defer c.Unlock()
	task := c.tasks[id]
	delete(c.tasks, id)
	task.Stop()
}

func (c *controller) StartAll() {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return
	}
	for _, task := range c.tasks {
		go task.Start()
	}
}

func (c *controller) StopAll() {
	c.Lock()
	defer c.Unlock()
	for _, task := range c.tasks {
		task.Stop()
	}
}

func (c *controller) WaitAll() {
	c.Lock()
	defer c.Unlock()
	for _, task := range c.tasks {
		task.Wait()
	}
}

func (c *controller) Close() {
	c.Lock()
	defer c.Unlock()
	c.closed = true
}
