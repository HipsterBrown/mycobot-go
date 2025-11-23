package protocol

// Protocol frame markers
const (
	Header byte = 0xFE
	Footer byte = 0xFA
)

// System status commands
const (
	SoftwareVersion      byte = 0x02
	GetRobotID          byte = 0x03
	SetRobotID          byte = 0x04
	PowerOn             byte = 0x10
	PowerOff            byte = 0x11
	IsPowerOn           byte = 0x12
	ReleaseAllServos    byte = 0x13
	IsControllerConnected byte = 0x14
	ReadNextError       byte = 0x15
	SetFreshMode        byte = 0x16
	GetFreshMode        byte = 0x17
	SetFreeMode         byte = 0x1A
	IsFreeMode          byte = 0x1B
)

// MDI mode commands
const (
	GetAngles           byte = 0x20
	SendAngle           byte = 0x21
	SendAngles          byte = 0x22
	GetCoords           byte = 0x23
	SendCoord           byte = 0x24
	SendCoords          byte = 0x25
	Pause               byte = 0x26
	IsPaused            byte = 0x27
	Resume              byte = 0x28
	Stop                byte = 0x29
	IsInPosition        byte = 0x2A
	IsMoving            byte = 0x2B
)

// JOG mode commands
const (
	JogAngle            byte = 0x30
	JogAbsolute         byte = 0x31
	JogCoord            byte = 0x32
	JogIncrement        byte = 0x33
	JogStop             byte = 0x34
)

// Encoder commands
const (
	SetEncoder          byte = 0x3A
	GetEncoder          byte = 0x3B
	SetEncoders         byte = 0x3C
	GetEncoders         byte = 0x3D
)

// Running status and settings
const (
	GetSpeed            byte = 0x40
	SetSpeed            byte = 0x41
	GetJointMinAngle    byte = 0x4A
	GetJointMaxAngle    byte = 0x4B
	SetJointMin         byte = 0x4C
	SetJointMax         byte = 0x4D
)

// Servo control
const (
	IsServoEnable       byte = 0x50
	IsAllServoEnable    byte = 0x51
	SetServoData        byte = 0x52
	GetServoData        byte = 0x53
	SetServoCalibration byte = 0x54
	ReleaseServo        byte = 0x56
	FocusServo          byte = 0x57
)

// Atom IO
const (
	SetColor            byte = 0x6A
	SetDigitalOutput    byte = 0xA0
	GetDigitalInput     byte = 0xA1
	SetPWMOutput        byte = 0xA2
	GetGripperValue     byte = 0x67
	SetGripperState     byte = 0x68
	SetGripperValue     byte = 0x66
	SetGripperIni       byte = 0x69
	IsGripperMoving     byte = 0x6B
)

// Basic IO
const (
	SetBasicOutput      byte = 0xA0
	GetBasicInput       byte = 0xA1
)

// Gripper extended
const (
	InitGripper              byte = 0x38
	SetGripperProtectCurrent byte = 0x39
	GetGripperProtectCurrent byte = 0x37
	SetHTSGripperTorque      byte = 0x35
	GetHTSGripperTorque      byte = 0x36
)
