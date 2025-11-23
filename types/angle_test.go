package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAngle_ValidateForJoint(t *testing.T) {
	tests := []struct {
		name        string
		angle       Angle
		joint       JointID
		model       Model
		expectError bool
	}{
		{"valid within range", Angle(45), Joint1, ModelMyCobot280, false},
		{"valid at min", Angle(-165), Joint1, ModelMyCobot280, false},
		{"valid at max", Angle(165), Joint1, ModelMyCobot280, false},
		{"below min", Angle(-170), Joint1, ModelMyCobot280, true},
		{"above max", Angle(170), Joint1, ModelMyCobot280, true},
		{"different joint", Angle(170), Joint6, ModelMyCobot280, false}, // Joint6 has different limits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.angle.ValidateForJoint(tt.joint, tt.model)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAngles_Validate(t *testing.T) {
	validAngles := Angles{0, 45, -90, 30, -45, 90}
	invalidLengthAngles := Angles{0, 45, -90} // only 3 angles
	invalidValueAngles := Angles{0, 45, -200, 30, -45, 90} // -200 out of range

	err := validAngles.Validate(6, ModelMyCobot280)
	assert.NoError(t, err)

	err = invalidLengthAngles.Validate(6, ModelMyCobot280)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 6 angles")

	err = invalidValueAngles.Validate(6, ModelMyCobot280)
	assert.Error(t, err)
}
