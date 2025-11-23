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
	ModelMyCobot280: {
		{MinAngle: -165, MaxAngle: 165}, // Joint 1
		{MinAngle: -165, MaxAngle: 165}, // Joint 2
		{MinAngle: -165, MaxAngle: 165}, // Joint 3
		{MinAngle: -165, MaxAngle: 165}, // Joint 4
		{MinAngle: -165, MaxAngle: 165}, // Joint 5
		{MinAngle: -175, MaxAngle: 175}, // Joint 6
	},
	ModelMyCobot320: {
		{MinAngle: -170, MaxAngle: 170}, // Joint 1
		{MinAngle: -137, MaxAngle: 137}, // Joint 2
		{MinAngle: -150, MaxAngle: 150}, // Joint 3
		{MinAngle: -145, MaxAngle: 145}, // Joint 4
		{MinAngle: -165, MaxAngle: 165}, // Joint 5
		{MinAngle: -180, MaxAngle: 180}, // Joint 6
	},
	ModelMechArm270: {
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -175, MaxAngle: 175},
	},
	ModelMyPalletizer260: {
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -90, MaxAngle: 90},
		{MinAngle: -90, MaxAngle: 90},
		{MinAngle: -165, MaxAngle: 165},
	},
}
