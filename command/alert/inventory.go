package alert

import (
	"sync"
)

const defaultNumAttempts int = 3

type inventory struct {
	alerts map[string]int
	mutex  *sync.RWMutex
}

func NewInventory() *inventory {
	return &inventory{
		alerts: make(map[string]int),
		mutex:  new(sync.RWMutex),
	}
}

func (i *inventory) register(id string) {
	i.mutex.Lock()
	if _, ok := i.alerts[id]; ok {
		return
	}
	i.alerts[id] = defaultNumAttempts
	i.mutex.Unlock()
}

func (i *inventory) deregister(id string) {
	i.mutex.Lock()
	delete(i.alerts, id)
	i.mutex.Unlock()
}

func (i *inventory) decrement(id string) {
	i.mutex.Lock()
	if v, ok := i.alerts[id]; ok {
		i.alerts[id] = v - 1
	}
	i.mutex.Unlock()
}

func (i *inventory) remaining(id string) int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	remaining := 0
	if v, ok := i.alerts[id]; ok {
		remaining = v
	}
	return remaining
}
