package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpeed_Constants(t *testing.T) {
	assert.Equal(t, Speed(0), SpeedMin)
	assert.Equal(t, Speed(25), SpeedSlow)
	assert.Equal(t, Speed(50), SpeedMedium)
	assert.Equal(t, Speed(75), SpeedFast)
	assert.Equal(t, Speed(100), SpeedMax)
}

func TestSpeed_Validate(t *testing.T) {
	tests := []struct {
		speed       Speed
		expectError bool
	}{
		{SpeedMin, false},
		{SpeedMax, false},
		{SpeedMedium, false},
		{Speed(50), false},
		{Speed(-1), true},
		{Speed(101), true},
	}

	for _, tt := range tests {
		err := tt.speed.Validate()
		if tt.expectError {
			assert.Error(t, err, "speed=%d", tt.speed)
		} else {
			assert.NoError(t, err, "speed=%d", tt.speed)
		}
	}
}
