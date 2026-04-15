package types

// Model represents a robot model type
type Model string

const (
	ModelMyCobot280      Model = "MyCobot280"
	ModelMyCobot320      Model = "MyCobot320"
	ModelMechArm270      Model = "MechArm270"
	ModelMyPalletizer260 Model = "MyPalletizer260"
)

// JointLimit defines min/max angle for a joint
type JointLimit struct {
	MinAngle float64
	MaxAngle float64
}

// getJointLimits returns joint limits for a specific model and joint
func getJointLimits(model Model, joint JointID) JointLimit {
	limits := modelJointLimits[model]
	if int(joint) <= len(limits) {
		return limits[joint.Index()]
	}
	// Default safe limits if not found
	return JointLimit{MinAngle: -165, MaxAngle: 165}
}

var modelJointLimits = map[Model][]JointLimit{
	ModelMechArm270: {
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -175, MaxAngle: 175},
	},
}
