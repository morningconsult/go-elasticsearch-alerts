package lock

import "sync"

type Lock struct {
	mutex *sync.RWMutex
	have  *bool
}

func NewLock() *Lock {
	return &Lock{
		mutex: new(sync.RWMutex),
		have:  new(bool),
	}
}

func (l *Lock) Acquired() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return *l.have
}

func (l *Lock) Set(b bool) {
	l.mutex.Lock()
	*l.have = b
	l.mutex.Unlock()
}
