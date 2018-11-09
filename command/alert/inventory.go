package alert

import (
	"sync"
)

const defaultNumAttempts int = 3

type inventory struct {
	alerts map[string]int
	mutex  sync.RWMutex
}

func (i *inventory) register(id string) {
	mutex.Lock()
	if _, ok := i.alerts[id]; ok {
		return
	}
	i.alerts[id] = defaultNumAttempts
	mutex.Unlock()
}

func (i *inventory) deregister(id string) {
	mutex.Lock()
	delete(i.alerts, id)
	mutex.Unlock()
}

func (i *inventory) decrement(id string) {
	mutex.Lock()
	if v, ok := i.alerts[id]; ok {
		i.alerts[id] = v - 1
	}
	mutex.Unlock()
}

func (i *inventory) remaining(id string) int {
	mutex.RLock()
	defer mutex.RUnlock()

	remaining := 0
	if v, ok := i.alerts[id]; ok {
		remaining = v
	}
	return remaining
}
