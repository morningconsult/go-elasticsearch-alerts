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

package alert

import (
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func TestInventory(t *testing.T) {
	active := newInventory()

	randomUUID := func() string {
		id, err := uuid.GenerateUUID()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}

	id := randomUUID()
	active.register(id)
	if _, ok := active.alerts[id]; !ok {
		t.Fatal("inventory.register() did not register new id")
	}

	remaining := active.remaining(id)
	if remaining != defaultNumAttempts {
		t.Fatalf("expected %d remaining attempts, got %d", defaultNumAttempts, remaining)
	}

	active.decrement(id)
	if d, ok := active.alerts[id]; !ok || d != (remaining-1) {
		t.Fatalf("inventory.deregister() did not decremeint count by 1")
	}
	active.deregister(id)
	if _, ok := active.alerts[id]; ok {
		t.Fatal("inventory.deregister() did not delete id from inventory")
	}
}
