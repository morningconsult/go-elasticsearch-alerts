// Copyright 2019 The Morning Consult, LLC or its affiliates. All Rights Reserved.
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

package lock

import "sync"

// Lock is used as a mutex to synchronize between nodes.
type Lock struct {
	mutex *sync.RWMutex
	have  *bool
}

// NewLock creates a new *Lock instance.
func NewLock() *Lock {
	return &Lock{
		mutex: new(sync.RWMutex),
		have:  new(bool),
	}
}

// Acquired returns true if the lock has been acquired and
// false otherwise.
func (l *Lock) Acquired() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return *l.have
}

// Set is used to set whether or not the lock has been
// acquired.
func (l *Lock) Set(b bool) {
	l.mutex.Lock()
	*l.have = b
	l.mutex.Unlock()
}
