package task

import (
	"log"
	"sync"
	"time"
)

var (
	taskId = uint(0)
)

type WorkFunc func(cancel <-chan interface{}) error

type Task interface {
	Start()
	Stop()
	Wait()
}

type task struct {
	// constructor args
	id   uint
	work WorkFunc
	// internal state
	sync.WaitGroup
	state        state
	backoff      uint
	backoffSteps []time.Duration
	cancel       chan interface{}
}

func CreateTask(id uint, work WorkFunc) Task {
	return &task{
		id,
		work,
		sync.WaitGroup{},
		Created,
		0,
		[]time.Duration{0, 5 * time.Second, 15 * time.Second, 30 * time.Second},
		make(chan interface{}),
	}
}

func (t *task) Start() {
	t.state = Running
	log.Println("starting task", t.id)
	for t.state == Running {
		t.Add(1)
		if t.backoff > 0 {
			b := t.backoffSteps[t.backoff]
			log.Printf("task %d backoff for %d seconds", t.id, int(b.Seconds()))
			select {
			case <-t.cancel:
				log.Printf("task %d is canceled during backoff", t.id)
			case <-time.After(t.backoffSteps[t.backoff]):
			}
		}
		if t.state == Running {
			log.Println("task", t.id, "starting work")
			if err := t.work(t.cancel); err != nil {
				log.Println("error during task", t.id, err)
				t.backoff = min(t.backoff+1, uint(len(t.backoffSteps)))
			} else {
				t.backoff = 0
			}
		}
		log.Println("task", t.id, "work ended")
		t.Done()
	}
}

func (t *task) Stop() {
	t.state = Stopped
	log.Printf("task#%d.state = Stopped", t.id)
	t.cancel <- nil
	log.Printf("task#%d.cancel <- nil", t.id)
}

func (t *task) Wait() {
	t.WaitGroup.Wait()
}
