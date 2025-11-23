package types

import "fmt"

// Angle represents a joint angle in degrees
type Angle float64

// ValidateForJoint checks if angle is valid for specific joint and model
func (a Angle) ValidateForJoint(joint JointID, model Model) error {
	limits := getJointLimits(model, joint)
	if float64(a) < limits.MinAngle || float64(a) > limits.MaxAngle {
		return fmt.Errorf("angle %.2f out of range [%.2f, %.2f] for joint %d on %s",
			a, limits.MinAngle, limits.MaxAngle, joint, model)
	}
	return nil
}

// Angles represents a set of joint angles
type Angles []Angle

// Validate checks if all angles are valid for the given model
func (a Angles) Validate(jointCount int, model Model) error {
	if len(a) != jointCount {
		return fmt.Errorf("expected %d angles, got %d", jointCount, len(a))
	}

	for i, angle := range a {
		joint := JointID(i + 1)
		if err := angle.ValidateForJoint(joint, model); err != nil {
			return err
		}
	}

	return nil
}

// ToFloat64 converts Angles to []float64
func (a Angles) ToFloat64() []float64 {
	result := make([]float64, len(a))
	for i, angle := range a {
		result[i] = float64(angle)
	}
	return result
}
