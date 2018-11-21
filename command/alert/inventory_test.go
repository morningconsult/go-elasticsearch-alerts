package alert

import (
	"testing"

	"github.com/hashicorp/go-uuid"
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
	if d, ok := active.alerts[id]; !ok || d != (remaining - 1) {
		t.Fatalf("inventory.deregister() did not decremeint count by 1")
	}
	active.deregister(id)
	if _, ok := active.alerts[id]; ok {
		t.Fatal("inventory.deregister() did not delete id from inventory")
	}

}