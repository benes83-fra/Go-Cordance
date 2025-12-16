package ecs

import (
	"testing"
)

type testComponent struct {
	updated bool
}

func (t *testComponent) Update(dt float32) {
	if dt >= 0 {
		t.updated = true
	}
}

func TestEntity_AddAndUpdate(t *testing.T) {
	e := NewEntity(42)
	tc := &testComponent{}
	e.AddComponent(tc)

	if len(e.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(e.Components))
	}

	e.Update(0.016)
	if !tc.updated {
		t.Fatalf("expected component to be updated")
	}
}
