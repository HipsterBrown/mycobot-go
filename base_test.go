package mycobot

import (
	"testing"
	"time"

	"github.com/hipsterbrown/mycobot-go/types"
)

func TestWithDefaultTimeout_setsField(t *testing.T) {
	b := newBase("/dev/null", getModelConfig(types.ModelMechArm270))
	WithDefaultTimeout(250 * time.Millisecond)(b)

	if b.defaultTimeout != 250*time.Millisecond {
		t.Fatalf("defaultTimeout = %v, want 250ms", b.defaultTimeout)
	}
}
