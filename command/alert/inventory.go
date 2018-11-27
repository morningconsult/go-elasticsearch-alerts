// Copyright 2018 The Morning Consult, LLC or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//         https://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package alert

import "sync"

const defaultNumAttempts int = 3

type inventory struct {
	alerts map[string]int
	mutex  *sync.RWMutex
}

func newInventory() *inventory {
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
