package types

import "fmt"

// JointID represents a robot joint identifier (1-based indexing)
type JointID int

const (
	Joint1 JointID = 1
	Joint2 JointID = 2
	Joint3 JointID = 3
	Joint4 JointID = 4
	Joint5 JointID = 5
	Joint6 JointID = 6
)

// Validate checks if joint ID is valid for given joint count
func (j JointID) Validate(jointCount int) error {
	if j < 1 || int(j) > jointCount {
		return fmt.Errorf("invalid joint ID %d for robot with %d joints", j, jointCount)
	}
	return nil
}

// Index returns 0-based index for array access
func (j JointID) Index() int {
	return int(j) - 1
}
