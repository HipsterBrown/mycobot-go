package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJointID_Constants(t *testing.T) {
	assert.Equal(t, JointID(1), Joint1)
	assert.Equal(t, JointID(2), Joint2)
	assert.Equal(t, JointID(3), Joint3)
	assert.Equal(t, JointID(4), Joint4)
	assert.Equal(t, JointID(5), Joint5)
	assert.Equal(t, JointID(6), Joint6)
}

func TestJointID_Validate(t *testing.T) {
	tests := []struct {
		joint       JointID
		jointCount  int
		expectError bool
	}{
		{Joint1, 6, false},
		{Joint6, 6, false},
		{JointID(0), 6, true},
		{JointID(7), 6, true},
		{Joint4, 4, false},
		{Joint5, 4, true},
	}

	for _, tt := range tests {
		err := tt.joint.Validate(tt.jointCount)
		if tt.expectError {
			assert.Error(t, err, "joint=%d, count=%d", tt.joint, tt.jointCount)
		} else {
			assert.NoError(t, err, "joint=%d, count=%d", tt.joint, tt.jointCount)
		}
	}
}
