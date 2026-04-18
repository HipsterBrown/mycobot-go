package types

// CoordAxis represents a coordinate axis for single-axis movement.
type CoordAxis int

const (
	AxisX CoordAxis = iota
	AxisY
	AxisZ
	AxisRx
	AxisRy
	AxisRz
)

// PositionFlag specifies whether IsInPosition checks angles or coordinates.
type PositionFlag int

const (
	PositionAngles PositionFlag = 0
	PositionCoords PositionFlag = 1
)
