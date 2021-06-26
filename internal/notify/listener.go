package notify

import (
	"sync"
	"time"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/lib/pq"
)

var Log logger.Logger = logger.Default

const conninfo = "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable"

type listener struct {
	sync.Mutex

	impl     *pq.Listener
	handlers map[string]func(string)

	closed bool
	close  chan interface{}
}

func (l *listener) Listen(channel string, handler func(payload string)) error {
	l.Lock()
	defer l.Unlock()
	l.handlers[channel] = handler
	err := l.impl.Listen(channel)
	if err != nil {
		delete(l.handlers, channel)
		return err
	}
	return nil
}

func (l *listener) Unlisten(channel string) {
	l.Lock()
	defer l.Unlock()
	l.impl.Unlisten(channel)
	delete(l.handlers, channel)
}

func (l *listener) Close() {
	l.Lock()
	defer l.Unlock()
	if l.closed {
		return
	}
	defer func() { l.closed = true }()
	l.close <- nil
	l.handlers = make(map[string]func(string))
	logger.Default.Info("l.impl.UnlistenAll()")
	l.impl.UnlistenAll()
	logger.Default.Info("l.impl.Close()")
	l.impl.Close()
}

func NewListener() *listener {
	conninfo, err := database.GetConninfo()
	if err != nil {
		Log.Warn("error reading connection info for postgresql listener: %s", err)
	}
	l := &listener{
		Mutex:    sync.Mutex{},
		impl:     pq.NewListener(conninfo, 5*time.Second, time.Minute, handleListenerEvent),
		handlers: make(map[string]func(string)),
		close:    make(chan interface{}),
		closed:   false,
	}
	go func() {
		for {
			select {
			case <-time.After(1 * time.Minute):
				go l.impl.Ping()
			case n := <-l.impl.NotificationChannel():
				if f, ok := l.handlers[n.Channel]; ok {
					f(n.Extra)
				}
			case <-l.close:
				Log.Info("Closing listener...")
				return
			}
		}
	}()
	return l
}

func handleListenerEvent(event pq.ListenerEventType, err error) {
	switch event {
	case pq.ListenerEventConnected:
		Log.Info("ListenerEventConnected: %s", err)
	case pq.ListenerEventDisconnected:
		Log.Info("ListenerEventDisconnected: %s", err)
	case pq.ListenerEventReconnected:
		Log.Info("ListenerEventReconnected: %s", err)
	case pq.ListenerEventConnectionAttemptFailed:
		Log.Info("ListenerEventConnectionAttemptFailed: %s", err)
	default:
		Log.Info("Unknown event type %v: %s", event, err)
	}
}
