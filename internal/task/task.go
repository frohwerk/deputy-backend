package task

import (
	"sync"
	"time"

	"github.com/frohwerk/deputy-backend/internal/logger"
)

var (
	taskId = uint(0)
)

type WorkFunc func(cancel <-chan interface{}) error

type Task interface {
	Name() string
	Start()
	Stop()
	Wait()
}

type task struct {
	// constructor args
	id   string
	log  logger.Logger
	work WorkFunc
	// internal state
	sync.WaitGroup
	state        state
	backoff      uint
	backoffSteps []time.Duration
	cancel       chan interface{}
}

func CreateTask(id string, log logger.Logger, work WorkFunc) Task {
	return &task{
		id,
		log,
		work,
		sync.WaitGroup{},
		Created,
		0,
		[]time.Duration{0, 1 * time.Second, 10 * time.Second, 60 * time.Second, 300 * time.Second},
		make(chan interface{}),
	}
}

func (t *task) Name() string {
	return t.id
}

func (t *task) Start() {
	t.state = Running
	t.log.Debug("starting task %s", t.id)
	for t.state == Running {
		t.Add(1)
		if t.backoff > 0 {
			b := t.backoffSteps[t.backoff]
			t.log.Trace("task %s backoff for %d seconds", t.id, int(b.Seconds()))
			select {
			case <-t.cancel:
				t.log.Info("task %s is canceled during backoff", t.id)
			case <-time.After(t.backoffSteps[t.backoff]):
			}
		}
		if t.state == Running {
			t.log.Debug("task %s starting work", t.id)
			if err := t.work(t.cancel); err != nil {
				t.log.Error("error during task %s: %s", t.id, err)
				t.backoff = min(t.backoff+1, uint(len(t.backoffSteps)-1))
			} else {
				t.backoff = 0
			}
		}
		t.log.Debug("task %s work ended", t.id)
		t.Done()
	}
}

func (t *task) Stop() {
	t.state = Stopped
	t.log.Trace("task#%s.state = Stopped", t.id)
	t.cancel <- nil
	t.log.Trace("task#%s.cancel <- nil", t.id)
}

func (t *task) Wait() {
	t.WaitGroup.Wait()
}
